package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestHandleAdminInitCommand_DiscordOpenPanel(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]
	store.schedules[date] = &repository.DaySchedule{Date: date, DayStatus: "celebration", AvailableMeals: []string{"lunch"}, Reason: "Company event"}

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord, Raw: events.APIGatewayV2HTTPRequest{}})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: json.RawMessage(`{"action":"open","date":"` + date + `"}`)})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 4 || body.Data == nil || body.Data.Flags != 64 {
		t.Fatalf("response = %#v, want ephemeral message", body)
	}
	if body.Data.Embeds[0].Title != adminInitDiscordPanelTitle {
		t.Fatalf("title = %q", body.Data.Embeds[0].Title)
	}
	if len(body.Data.Components) != 5 {
		t.Fatalf("components = %d, want 5 rows", len(body.Data.Components))
	}
	if len(body.Data.Components[0].Components) != 1 || body.Data.Components[0].Components[0].Label != "Edit Details" {
		t.Fatalf("first row = %#v, want only Edit Details", body.Data.Components[0].Components)
	}
	if !strings.Contains(body.Data.Embeds[0].Fields[3].Value, "Company event") {
		t.Fatalf("reason field = %q", body.Data.Embeds[0].Fields[3].Value)
	}
}

func TestDiscordComponentPayload_AdminInitEditOpensModal(t *testing.T) {
	draft := adminInitScheduleDraft{Date: "2026-05-15", EndDate: "2026-05-18", UseRange: true, Status: "normal", Meals: []string{"lunch"}, Reason: "Ops"}
	commandName, raw, err := discordComponentPayload(interactionBody{Data: interactionData{CustomID: discordAdminInitCustomID(adminInitDiscordActionEdit, draft)}, Message: buildDiscordAdminInitMessage(draft, "", discord.NoticeToneInfo)})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "admin-init" {
		t.Fatalf("commandName = %q, want admin-init", commandName)
	}
	var opts payload.AdminInitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Action != adminInitDiscordActionEdit || opts.Reason != "Ops" || !opts.UseRange {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestHandleAdminInitCommand_DiscordEditReturnsModal(t *testing.T) {
	draft := adminInitScheduleDraft{Date: "2026-05-15", EndDate: "2026-05-18", UseRange: true, Status: "normal", Meals: []string{"lunch"}, Reason: "Ops"}
	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleAdminInitCommand(ctx, newAdminInitTestStore(), nil, mustInitDateParser(t), payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: mustJSON(t, discordAdminInitOptionsFromDraft(adminInitDiscordActionEdit, draft))})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}

	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 9 || body.Data == nil || body.Data.Title != "Edit Admin Setup" {
		t.Fatalf("modal response = %#v", body)
	}
	if len(body.Data.Components) != 3 {
		t.Fatalf("modal components = %d, want 3", len(body.Data.Components))
	}
	startInput := body.Data.Components[0].Components[0]
	endInput := body.Data.Components[1].Components[0]
	reasonInput := body.Data.Components[2].Components[0]
	if startInput.Required == nil || !*startInput.Required {
		t.Fatalf("start required = %#v, want true", startInput.Required)
	}
	if endInput.Required == nil || *endInput.Required {
		t.Fatalf("end required = %#v, want false", endInput.Required)
	}
	if reasonInput.Required == nil || *reasonInput.Required {
		t.Fatalf("reason required = %#v, want false", reasonInput.Required)
	}
}

func TestDiscordComponentPayload_AdminInitModalSubmit(t *testing.T) {
	draft := adminInitScheduleDraft{Date: "2026-05-15", UseRange: true, Status: "normal", Meals: []string{"lunch"}}
	commandName, raw, err := discordComponentPayload(interactionBody{Type: 5, Data: interactionData{CustomID: discordAdminInitCustomID(adminInitDiscordActionModalSave, draft), Components: []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "date", Value: "2026-05-20"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "end_date", Value: "2026-05-22"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "reason", Value: "Bulk ops"}}},
	}}})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	if commandName != "admin-init" {
		t.Fatalf("commandName = %q", commandName)
	}
	var opts payload.AdminInitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.Action != adminInitDiscordActionModalSave || opts.Date != "2026-05-20" || opts.EndDate != "2026-05-22" || opts.Reason != "Bulk ops" {
		t.Fatalf("opts = %+v", opts)
	}
	if !opts.UseRange {
		t.Fatalf("UseRange = false, want true for filled end date")
	}
}

func TestDiscordComponentPayload_AdminInitModalSubmitEmptyEndDateIsSingleDate(t *testing.T) {
	draft := adminInitScheduleDraft{Date: "2026-05-15", EndDate: "2026-05-18", Status: "normal", Meals: []string{"lunch"}}
	_, raw, err := discordComponentPayload(interactionBody{Type: 5, Data: interactionData{CustomID: discordAdminInitCustomID(adminInitDiscordActionModalSave, draft), Components: []discord.Component{
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "date", Value: "2026-05-20"}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "end_date", Value: ""}}},
		{Type: discord.ComponentTypeActionRow, Components: []discord.Component{{CustomID: "reason", Value: "Single ops"}}},
	}}})
	if err != nil {
		t.Fatalf("discordComponentPayload() error = %v", err)
	}
	var opts payload.AdminInitOptions
	if err := json.Unmarshal(raw, &opts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if opts.UseRange || opts.EndDate != "" {
		t.Fatalf("opts = %+v, want single-date mode", opts)
	}
}

func TestHandleAdminInitCommand_DiscordSaveSingleDate(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: json.RawMessage(`{"action":"save","date":"` + date + `","status":"normal","meals":["lunch","snacks"],"reason":"Ops"}`)})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	schedule := store.schedules[date]
	if schedule == nil || schedule.DayStatus != "normal" || strings.Join(schedule.AvailableMeals, ",") != "lunch,snacks" || schedule.Reason != "Ops" {
		t.Fatalf("schedule = %+v", schedule)
	}
	var body RouterResponse
	_ = json.Unmarshal([]byte(recorder.finalResponse().Body), &body)
	if body.Type != 7 || body.Data.Embeds[0].Color != discord.SuccessColor {
		t.Fatalf("response = %#v", body)
	}
}

func TestHandleAdminInitCommand_DiscordOfficeClosedClearsMeals(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	date := nextInitTestDates(t, dateParser, 1)[0]

	ctx, _ := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: json.RawMessage(`{"action":"save","date":"` + date + `","status":"office_closed","meals":["lunch"],"reason":"Maintenance"}`)})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	if got := store.schedules[date]; got == nil || got.DayStatus != "office_closed" || len(got.AvailableMeals) != 0 {
		t.Fatalf("schedule = %+v", got)
	}
}

func TestHandleAdminInitCommand_DiscordSaveRangeFromEndDate(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	start, end, weekdays := nextAdminInitWeekdayRange(t, dateParser, 2)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: json.RawMessage(`{"action":"save","date":"` + start + `","end_date":"` + end + `","status":"normal","meals":["lunch"],"reason":"Bulk ops"}`)})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	for _, date := range weekdays {
		schedule := store.schedules[date]
		if schedule == nil || schedule.DayStatus != "normal" || strings.Join(schedule.AvailableMeals, ",") != "lunch" || schedule.Reason != "Bulk ops" {
			t.Fatalf("schedule[%s] = %+v", date, schedule)
		}
	}
	var body RouterResponse
	_ = json.Unmarshal([]byte(recorder.finalResponse().Body), &body)
	if body.Type != 7 || body.Data.Embeds[0].Color != discord.SuccessColor {
		t.Fatalf("response = %#v", body)
	}
}

func TestHandleAdminInitCommand_DiscordRangeEndDateMustBeAfterStartDate(t *testing.T) {
	store := newAdminInitTestStore()
	dateParser := mustInitDateParser(t)
	start, _, _ := nextAdminInitWeekdayRange(t, dateParser, 2)

	ctx, recorder := withReplyRecorder(context.Background(), HandlerRequest{Platform: PlatformDiscord})
	err := handleAdminInitCommand(ctx, store, nil, dateParser, payload.CommandEvent{UserID: "admin1", Role: "admin", Source: "discord", Options: json.RawMessage(`{"action":"save","date":"` + start + `","end_date":"` + start + `","status":"normal","meals":["lunch"]}`)})
	if err != nil {
		t.Fatalf("handleAdminInitCommand() error = %v", err)
	}
	if len(store.schedules) != 0 {
		t.Fatalf("schedules saved on validation failure: %+v", store.schedules)
	}
	var body RouterResponse
	if err := json.Unmarshal([]byte(recorder.finalResponse().Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Type != 7 || body.Data.Embeds[0].Color != discord.WarningColor || !strings.Contains(body.Data.Embeds[0].Description, "end date must be after start date") {
		t.Fatalf("response = %#v", body)
	}
}

func TestDiscordAdminInitCustomID_StaysWithinDiscordLimit(t *testing.T) {
	draft := adminInitScheduleDraft{Date: "2026-05-15", EndDate: "2026-05-22", UseRange: true, Status: "office_closed", Meals: []string{"lunch", "snacks", "iftar", "event_dinner", "optional_dinner"}, Reason: strings.Repeat("x", 200)}
	id := discordAdminInitCustomID(adminInitDiscordActionMeal+"optional_dinner", draft)
	if len(id) > 100 {
		t.Fatalf("custom id length = %d, want <= 100: %q", len(id), id)
	}
}

func mustJSON(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return raw
}
