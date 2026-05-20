package gchat

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseMealArgs_Valid(t *testing.T) {
	opts, err := parseMealArgs("in lunch 2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "in" {
		t.Errorf("status = %q, want %q", opts["status"], "in")
	}
	if opts["meal"] != "lunch" {
		t.Errorf("meal = %q, want %q", opts["meal"], "lunch")
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestParseMealArgs_StatusOnly(t *testing.T) {
	opts, err := parseMealArgs("out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "out" {
		t.Errorf("status = %q, want %q", opts["status"], "out")
	}
	if opts["meal"] != "all" {
		t.Errorf("meal = %q, want %q", opts["meal"], "all")
	}
}

func TestParseMealArgs_ToggleMealOnly(t *testing.T) {
	opts, err := parseMealArgs("lunch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "" {
		t.Errorf("status = %q, want empty", opts["status"])
	}
	if opts["meal"] != "lunch" {
		t.Errorf("meal = %q, want %q", opts["meal"], "lunch")
	}
	if opts["date"] != "" {
		t.Errorf("date = %q, want empty", opts["date"])
	}
}

func TestParseMealArgs_ToggleDateOnly(t *testing.T) {
	opts, err := parseMealArgs("2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "" {
		t.Errorf("status = %q, want empty", opts["status"])
	}
	if opts["meal"] != "all" {
		t.Errorf("meal = %q, want %q", opts["meal"], "all")
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestParseMealArgs_InvalidMealType(t *testing.T) {
	_, err := parseMealArgs("in brunch")
	if err == nil {
		t.Fatal("expected error for invalid meal type, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid meal_type") {
		t.Errorf("error = %q, want invalid meal_type message", err.Error())
	}
}

func TestParseMealArgs_InvalidDateFormat(t *testing.T) {
	opts, err := parseMealArgs("in lunch 20-03-2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter now passes raw date string - Lambda will validate
	if opts["date"] != "20-03-2026" {
		t.Errorf("date = %q, want %q", opts["date"], "20-03-2026")
	}
}

func TestParseMealArgs_Empty(t *testing.T) {
	opts, err := parseMealArgs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["status"] != "" {
		t.Errorf("status = %q, want empty", opts["status"])
	}
	if opts["meal"] != "all" {
		t.Errorf("meal = %q, want %q", opts["meal"], "all")
	}
}

func TestParseLocationArgs_ToggleDateOnly(t *testing.T) {
	opts := parseLocationArgs("2026-03-20")
	if opts["location"] != "" {
		t.Errorf("location = %q, want empty", opts["location"])
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestParseLocationArgs_ToggleDefaultDate(t *testing.T) {
	opts := parseLocationArgs("")
	if opts["location"] != "" {
		t.Errorf("location = %q, want empty", opts["location"])
	}
	if opts["date"] != "" {
		t.Errorf("date = %q, want empty", opts["date"])
	}
}

func TestParseHeadcountArgs_MissingDate(t *testing.T) {
	opts, err := parseHeadcountArgs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter passes empty string - Lambda will default to tomorrow
	if opts["date"] != "" {
		t.Errorf("date = %q, want empty string", opts["date"])
	}
}

func TestParseHeadcountArgs_InvalidDateFormat(t *testing.T) {
	opts, err := parseHeadcountArgs("20-03-2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Adapter now passes raw date string - Lambda will validate
	if opts["date"] != "20-03-2026" {
		t.Errorf("date = %q, want %q", opts["date"], "20-03-2026")
	}
}

func TestParseHeadcountArgs_Valid(t *testing.T) {
	opts, err := parseHeadcountArgs("2026-03-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["date"] != "2026-03-20" {
		t.Errorf("date = %q, want %q", opts["date"], "2026-03-20")
	}
}

func TestCommandMapping(t *testing.T) {
	expected := map[int64]string{
		1:  "meal",
		2:  "location",
		3:  "team-summary",
		4:  "headcount",
		5:  "status",
		6:  "schedule-day",
		7:  "override",
		9:  "help",
		10: "init",
		11: "admin-init",
	}
	for id, want := range expected {
		got, ok := gchatCommandNames[id]
		if !ok {
			t.Errorf("command ID %d not found in mapping", id)
			continue
		}
		if got != want {
			t.Errorf("command ID %d = %q, want %q", id, got, want)
		}
	}
}

func TestToCommandEvent_AdminInitOpen(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User: Sender{Name: "users/123"},
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 11},
				Space:              Space{Name: "spaces/abc"},
				Message:            &Message{ArgumentText: "2026-05-20"},
			},
		},
	}

	ce, err := ToCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "admin-init" || ce.Source != "gchat" {
		t.Fatalf("unexpected command event: %+v", ce)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts["action"] != "open" || opts["date"] != "2026-05-20" {
		t.Fatalf("unexpected admin-init options: %#v", opts)
	}
}

func TestToCommandEvent_InitOpen(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User: Sender{Name: "users/123"},
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 10},
				Space:              Space{Name: "spaces/abc"},
				Message:            &Message{ArgumentText: "tomorrow"},
			},
		},
	}

	ce, err := ToCommandEvent(evt, "user1", "employee")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "init" {
		t.Fatalf("CommandName = %q, want init", ce.CommandName)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts["action"] != "open" || opts["date"] != "tomorrow" {
		t.Fatalf("unexpected init options: %#v", opts)
	}
}

func TestToCardActionCommandEvent_InitSave(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User:                 Sender{Name: "users/123"},
			Space:                Space{Name: "spaces/abc"},
			ButtonClickedPayload: &ButtonClickedPayload{},
		},
		CommonEventObject: CommonEventObject{
			Parameters: map[string]string{"action": InitCardFunctionSave, "date": "2026-05-15"},
			FormInputs: map[string]FormInput{
				"dates":    {StringInputs: &StringInputs{Value: []string{"2026-05-15", "2026-05-16"}}},
				"location": {StringInputs: &StringInputs{Value: []string{"wfh"}}},
				"meals":    {StringInputs: &StringInputs{Value: []string{"lunch"}}},
			},
		},
	}

	ce, err := ToCardActionCommandEvent(evt, "user1", "employee")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "init" || ce.Source != "gchat" {
		t.Fatalf("unexpected command event: %+v", ce)
	}
	var opts struct {
		Action   string   `json:"action"`
		Date     string   `json:"date"`
		Dates    []string `json:"dates"`
		Location string   `json:"location"`
		Meals    []string `json:"meals"`
	}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts.Action != "apply" || opts.Date != "2026-05-15" || opts.Location != "wfh" {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if strings.Join(opts.Dates, ",") != "2026-05-15,2026-05-16" {
		t.Fatalf("dates = %#v", opts.Dates)
	}
	if strings.Join(opts.Meals, ",") != "lunch" {
		t.Fatalf("meals = %#v", opts.Meals)
	}
}

func TestToCardActionCommandEvent_InitEdit(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User:                 Sender{Name: "users/123"},
			Space:                Space{Name: "spaces/abc"},
			ButtonClickedPayload: &ButtonClickedPayload{},
		},
		CommonEventObject: CommonEventObject{
			Parameters: map[string]string{"action": InitCardFunctionEdit, "date": "2026-05-15"},
		},
	}

	ce, err := ToCardActionCommandEvent(evt, "user1", "employee")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "init" || ce.Source != "gchat" {
		t.Fatalf("unexpected command event: %+v", ce)
	}
	var opts struct {
		Action string `json:"action"`
		Date   string `json:"date"`
	}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts.Action != "open" || opts.Date != "2026-05-15" {
		t.Fatalf("unexpected options: %+v", opts)
	}
}

func TestToCardActionCommandEvent_AdminInitScheduleSave(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User:                 Sender{Name: "users/123"},
			Space:                Space{Name: "spaces/abc"},
			ButtonClickedPayload: &ButtonClickedPayload{},
		},
		CommonEventObject: CommonEventObject{
			Parameters: map[string]string{"action": AdminInitFunctionScheduleSave, "date": "2026-05-20"},
			FormInputs: map[string]FormInput{
				"date":       {StringInputs: &StringInputs{Value: []string{"2026-05-21"}}},
				"date_range": {StringInputs: &StringInputs{Value: []string{"true"}}},
				"status":     {StringInputs: &StringInputs{Value: []string{"celebration"}}},
				"meals":      {StringInputs: &StringInputs{Value: []string{"lunch", "snacks"}}},
				"reason":     {StringInputs: &StringInputs{Value: []string{"Company event"}}},
			},
		},
	}

	ce, err := ToCardActionCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "admin-init" || ce.Source != "gchat" {
		t.Fatalf("unexpected command event: %+v", ce)
	}
	var opts struct {
		Action   string   `json:"action"`
		Date     string   `json:"date"`
		UseRange bool     `json:"use_range"`
		Status   string   `json:"status"`
		Meals    []string `json:"meals"`
		Reason   string   `json:"reason"`
	}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts.Action != "schedule_apply" || opts.Date != "2026-05-21" || opts.UseRange || opts.Status != "celebration" || opts.Reason != "Company event" {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if strings.Join(opts.Meals, ",") != "lunch,snacks" {
		t.Fatalf("meals = %#v", opts.Meals)
	}
}

func TestToCardActionCommandEvent_AdminInitSingleDateToggle(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User:  Sender{Name: "users/123"},
			Space: Space{Name: "spaces/abc"},
		},
		CommonEventObject: CommonEventObject{
			Parameters: map[string]string{"action": AdminInitFunctionRangeToggle, "date": "2026-05-20"},
			FormInputs: map[string]FormInput{
				"date":       {StringInputs: &StringInputs{Value: []string{"2026-05-21"}}},
				"end_date":   {StringInputs: &StringInputs{Value: []string{"2026-05-25"}}},
				"date_range": {StringInputs: &StringInputs{Value: []string{"true"}}},
				"status":     {StringInputs: &StringInputs{Value: []string{"normal"}}},
				"meals":      {StringInputs: &StringInputs{Value: []string{"lunch"}}},
				"reason":     {StringInputs: &StringInputs{Value: []string{"Range ops"}}},
			},
		},
	}

	ce, err := ToCardActionCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "admin-init" || ce.Source != "gchat" {
		t.Fatalf("unexpected command event: %+v", ce)
	}
	var opts struct {
		Action   string   `json:"action"`
		Date     string   `json:"date"`
		EndDate  string   `json:"end_date"`
		UseRange bool     `json:"use_range"`
		Status   string   `json:"status"`
		Meals    []string `json:"meals"`
		Reason   string   `json:"reason"`
	}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts.Action != "schedule_range_toggle" || opts.Date != "2026-05-21" || opts.EndDate != "2026-05-25" || opts.UseRange || opts.Status != "normal" || opts.Reason != "Range ops" {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if strings.Join(opts.Meals, ",") != "lunch" {
		t.Fatalf("meals = %#v", opts.Meals)
	}
}

func TestToCardActionCommandEvent_AdminInitRangeToggleUncheckedUsesRangeMode(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User:  Sender{Name: "users/123"},
			Space: Space{Name: "spaces/abc"},
		},
		CommonEventObject: CommonEventObject{
			Parameters: map[string]string{"action": AdminInitFunctionRangeToggle, "date": "2026-05-20", "use_range": "true"},
			FormInputs: map[string]FormInput{
				"date": {StringInputs: &StringInputs{Value: []string{"2026-05-21"}}},
			},
		},
	}

	ce, err := ToCardActionCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var opts struct {
		UseRange bool `json:"use_range"`
	}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if !opts.UseRange {
		t.Fatalf("use_range = false, want true when Mark a single date is absent from form inputs")
	}
}

func TestToCommandEvent_NilPayload(t *testing.T) {
	evt := Event{Chat: ChatEvent{}}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for nil payload, got nil")
	}
}

func TestToCommandEvent_UnknownCommandID(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 99},
			},
		},
	}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for unknown command ID, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command ID") {
		t.Errorf("error = %q, want unknown command ID message", err.Error())
	}
}

func TestToCommandEvent_MealParseError(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 1},
				Message:            &Message{ArgumentText: "lunch tomorrow extra"},
			},
		},
	}
	_, err := ToCommandEvent(evt, "user1", "member")
	if err == nil {
		t.Fatal("expected error for invalid meal status, got nil")
	}
}

func TestToCommandEvent_HeadcountParseError(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 4},
				Message:            &Message{ArgumentText: "invalid-date-format"},
			},
		},
	}
	ce, err := ToCommandEvent(evt, "user1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if opts["date"] != "" {
		t.Errorf("date = %q, want empty string", opts["date"])
	}
}

func TestToCommandEvent_HelpNoArgs(t *testing.T) {
	evt := Event{
		Chat: ChatEvent{
			User: Sender{Name: "users/123"},
			AppCommandPayload: &AppCommandPayload{
				AppCommandMetadata: AppCommandMetadata{AppCommandID: 9},
				Space:              Space{Name: "spaces/abc"},
			},
		},
	}

	ce, err := ToCommandEvent(evt, "user1", "employee")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce.CommandName != "help" {
		t.Fatalf("CommandName = %q, want help", ce.CommandName)
	}
	var opts map[string]interface{}
	if err := json.Unmarshal(ce.Options, &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options, got %#v", opts)
	}
}

func TestParseOverrideArgs_MealToggle(t *testing.T) {
	opts, err := parseOverrideArgs("alice@example.com meal tomorrow lunch -- Forgot to update")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["target"] != "alice@example.com" {
		t.Fatalf("target = %v, want alice@example.com", opts["target"])
	}
	if opts["entry"] != "meal" {
		t.Fatalf("entry = %v, want meal", opts["entry"])
	}
	if opts["meal"] != "lunch" {
		t.Fatalf("meal = %v, want lunch", opts["meal"])
	}
	if opts["value"] != "" {
		t.Fatalf("value = %v, want empty", opts["value"])
	}
	if opts["reason"] != "Forgot to update" {
		t.Fatalf("reason = %v, want reason text", opts["reason"])
	}
}

func TestParseOverrideArgs_LocationExplicitValue(t *testing.T) {
	opts, err := parseOverrideArgs("alice@example.com location 2026-05-10 wfh -- Doctor appointment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts["entry"] != "location" {
		t.Fatalf("entry = %v, want location", opts["entry"])
	}
	if opts["value"] != "wfh" {
		t.Fatalf("value = %v, want wfh", opts["value"])
	}
	if opts["reason"] != "Doctor appointment" {
		t.Fatalf("reason = %v, want reason text", opts["reason"])
	}
}

func TestParseOverrideArgs_MissingReason(t *testing.T) {
	_, err := parseOverrideArgs("alice@example.com location tomorrow")
	if err == nil {
		t.Fatal("expected error for missing reason")
	}
}

func TestParseOverrideArgs_AmbiguousMealWithoutSeparator(t *testing.T) {
	_, err := parseOverrideArgs("alice@example.com meal tomorrow lunch Forgot to update")
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "Ambiguous override syntax") {
		t.Fatalf("error = %q, want ambiguity guidance", err.Error())
	}
}

func TestParseOverrideArgs_AmbiguousLocationWithoutSeparator(t *testing.T) {
	_, err := parseOverrideArgs("alice@example.com location tomorrow office relocation needed")
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "Ambiguous override syntax") {
		t.Fatalf("error = %q, want ambiguity guidance", err.Error())
	}
}

func TestParseScheduleDayArgs_GovtHolidayReasonOnly(t *testing.T) {
	opts := parseScheduleDayArgs("2026-03-26 govt_holiday Independence Day")
	if opts["date"] != "2026-03-26" {
		t.Fatalf("date = %v, want 2026-03-26", opts["date"])
	}
	if opts["status"] != "govt_holiday" {
		t.Fatalf("status = %v, want govt_holiday", opts["status"])
	}
	if _, ok := opts["meals"]; ok {
		t.Fatalf("expected no meals option, got %v", opts["meals"])
	}
	if opts["reason"] != "Independence Day" {
		t.Fatalf("reason = %v, want Independence Day", opts["reason"])
	}
}

func TestParseScheduleDayArgs_OfficeClosedReasonOnly(t *testing.T) {
	opts := parseScheduleDayArgs("2026-03-28 office_closed Maintenance Window")
	if opts["status"] != "office_closed" {
		t.Fatalf("status = %v, want office_closed", opts["status"])
	}
	if _, ok := opts["meals"]; ok {
		t.Fatalf("expected no meals option, got %v", opts["meals"])
	}
	if opts["reason"] != "Maintenance Window" {
		t.Fatalf("reason = %v, want Maintenance Window", opts["reason"])
	}
}

func TestParseScheduleDayArgs_CelebrationMealsAndReason(t *testing.T) {
	opts := parseScheduleDayArgs("2026-03-27 celebration lunch,snacks Company Anniversary")
	if opts["meals"] != "lunch,snacks" {
		t.Fatalf("meals = %v, want lunch,snacks", opts["meals"])
	}
	if opts["reason"] != "Company Anniversary" {
		t.Fatalf("reason = %v, want Company Anniversary", opts["reason"])
	}
}

func TestParseTeamSummaryArgs_AdminTeamID(t *testing.T) {
	opts := parseTeamSummaryArgs("tomorrow team-42")
	if opts["team_id"] != "team-42" {
		t.Fatalf("team_id = %v, want team-42", opts["team_id"])
	}
	if opts["date"] != "tomorrow" {
		t.Fatalf("date = %v, want tomorrow", opts["date"])
	}
}

func TestParseTeamSummaryArgs_TeamIDOnly(t *testing.T) {
	opts := parseTeamSummaryArgs("team-42")
	if opts["team_id"] != "team-42" {
		t.Fatalf("team_id = %v, want team-42", opts["team_id"])
	}
	if opts["date"] != "" {
		t.Fatalf("date = %v, want empty", opts["date"])
	}
}

func TestParseTeamSummaryArgs_Empty(t *testing.T) {
	opts := parseTeamSummaryArgs("")
	if opts["date"] != "" {
		t.Fatalf("date = %v, want empty", opts["date"])
	}
	if opts["team_id"] != "" {
		t.Fatalf("team_id = %v, want empty", opts["team_id"])
	}
}

func TestParseTeamSummaryArgs_DateOnly(t *testing.T) {
	opts := parseTeamSummaryArgs("2026-05-10")
	if opts["date"] != "2026-05-10" {
		t.Fatalf("date = %v, want 2026-05-10", opts["date"])
	}
	if opts["team_id"] != "" {
		t.Fatalf("team_id = %v, want empty", opts["team_id"])
	}
}
