package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handleManagementCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	switch event.CommandName {
	case "team-summary":
		return handleTeamSummaryCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	case "override":
		return handleOverrideCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	default:
		return sendWarningReply(ctx, deps.cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

type overrideStore interface {
	ListActiveUsers(ctx context.Context) ([]repository.User, error)
	FindTeamsByLeadID(ctx context.Context, leadUserID string) ([]repository.Team, error)
	GetTeamMembers(ctx context.Context, teamID string) ([]repository.TeamMember, error)
	GetDay(ctx context.Context, date string) (*repository.DaySchedule, error)
	GetAvailableMeals(ctx context.Context, date string) ([]string, error)
	GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]repository.MealParticipation, error)
	GetParticipationsByDate(ctx context.Context, date string) ([]repository.MealParticipation, error)
	UpsertParticipation(ctx context.Context, p repository.MealParticipation) error
	GetWorkLocation(ctx context.Context, userID, date string) (*repository.WorkLocation, error)
	GetWorkLocationsByDate(ctx context.Context, date string) ([]repository.WorkLocation, error)
	UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation) error
	WriteAuditEntry(ctx context.Context, entry repository.AuditEntry) error
}

type overrideResult struct {
	TargetEmail string
	Entry       string
	Date        string
	Summary     string
	Reason      string
}

type overrideUserError struct {
	msg string
}

func (e *overrideUserError) Error() string {
	return e.msg
}

func userOverrideError(msg string) error {
	return &overrideUserError{msg: msg}
}

func handleOverrideCommand(ctx context.Context, store overrideStore, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	result, err := executeOverrideCommand(ctx, store, dateParser, event)
	if err != nil {
		var userErr *overrideUserError
		if errors.As(err, &userErr) {
			return sendWarningReply(ctx, cfg, event, userErr.Error())
		}
		return sendErrorReply(ctx, cfg, event, "Something went wrong saving the override. Please try again later.")
	}
	return sendReply(ctx, cfg, event, formatOverrideSuccess(result))
}

func executeOverrideCommand(ctx context.Context, store overrideStore, dateParser *dateutil.DateParser, event payload.CommandEvent) (*overrideResult, error) {
	if event.Role != "team_lead" && event.Role != "admin" {
		return nil, userOverrideError("You do not have permission to use `/override`.")
	}

	var opts payload.OverrideOptions
	if err := event.ParseOptions(&opts); err != nil {
		return nil, userOverrideError("Invalid command options.")
	}
	opts.Target = strings.TrimSpace(opts.Target)
	opts.Entry = strings.ToLower(strings.TrimSpace(opts.Entry))
	opts.Date = strings.TrimSpace(opts.Date)
	opts.Meal = strings.ToLower(strings.TrimSpace(opts.Meal))
	opts.Value = strings.ToLower(strings.TrimSpace(opts.Value))
	opts.Reason = strings.TrimSpace(opts.Reason)

	if opts.Target == "" {
		return nil, userOverrideError("Target email is required.")
	}
	if opts.Entry != "meal" && opts.Entry != "location" {
		return nil, userOverrideError("Entry must be `meal` or `location`.")
	}
	if opts.Reason == "" {
		return nil, userOverrideError("Override reason is required.")
	}

	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return nil, userOverrideError(fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}

	target, err := resolveOverrideTargetByEmail(ctx, store, opts.Target)
	if err != nil {
		return nil, err
	}
	if err := authorizeOverrideTarget(ctx, store, event, target.ID); err != nil {
		return nil, err
	}

	if opts.Entry == "meal" {
		return executeMealOverride(ctx, store, event, target, date, opts)
	}
	return executeLocationOverride(ctx, store, event, target, date, opts)
}

func resolveOverrideTargetByEmail(ctx context.Context, store overrideStore, targetEmail string) (*repository.User, error) {
	users, err := store.ListActiveUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Active && strings.EqualFold(user.Email, strings.TrimSpace(targetEmail)) {
			return &user, nil
		}
	}
	return nil, userOverrideError("Target user not found or inactive.")
}

func authorizeOverrideTarget(ctx context.Context, store overrideStore, event payload.CommandEvent, targetUserID string) error {
	if event.Role == "admin" {
		return nil
	}
	teams, err := store.FindTeamsByLeadID(ctx, event.UserID)
	if err != nil {
		return err
	}
	for _, team := range teams {
		members, err := store.GetTeamMembers(ctx, team.ID)
		if err != nil {
			return err
		}
		for _, member := range members {
			if member.UserID == targetUserID {
				return nil
			}
		}
	}
	return userOverrideError("You can only override entries for members of teams you lead.")
}

func executeMealOverride(ctx context.Context, store overrideStore, event payload.CommandEvent, target *repository.User, date string, opts payload.OverrideOptions) (*overrideResult, error) {
	meal := opts.Meal
	if meal == "" {
		meal = "all"
	}
	if !repository.IsValidMealOrAll(meal) {
		return nil, userOverrideError(fmt.Sprintf("`%s` is not a valid meal type. Choose from: lunch, snacks, iftar, event_dinner, optional_dinner, all.", meal))
	}
	if opts.Value != "" && opts.Value != "in" && opts.Value != "out" {
		return nil, userOverrideError("Meal override value must be `in` or `out`, or omitted to toggle.")
	}

	schedule, err := store.GetDay(ctx, date)
	if err != nil {
		return nil, err
	}
	if schedule != nil && (schedule.DayStatus == "office_closed" || schedule.DayStatus == "govt_holiday") {
		return nil, userOverrideError(fmt.Sprintf("Office is closed on %s - no meals are available.", date))
	}
	availableMeals, err := store.GetAvailableMeals(ctx, date)
	if err != nil {
		return nil, err
	}
	if len(availableMeals) == 0 {
		return nil, userOverrideError(fmt.Sprintf("No meals are configured for %s.", date))
	}

	statuses, err := services.GetUserMealStatus(ctx, store, store, target.ID, date)
	if err != nil {
		return nil, err
	}
	currentRecords, err := store.GetParticipationsByUserDate(ctx, target.ID, date)
	if err != nil {
		return nil, err
	}
	byMeal := make(map[string]repository.MealParticipation, len(currentRecords))
	for _, record := range currentRecords {
		byMeal[record.MealType] = record
	}

	targetStatuses := statuses
	if meal != "all" {
		found := false
		for _, status := range statuses {
			if status.MealType == meal {
				targetStatuses = []services.ResolvedStatus{status}
				found = true
				break
			}
		}
		if !found {
			return nil, userOverrideError(fmt.Sprintf("That meal is not available on %s.", date))
		}
	}

	changed := make([]string, 0, len(targetStatuses))
	for _, status := range targetStatuses {
		newValue := opts.Value == "in"
		if opts.Value == "" {
			newValue = status.Status != "opted_in"
		}

		existing, existed := byMeal[status.MealType]
		newRecord := repository.MealParticipation{
			UserID:   target.ID,
			Date:     date,
			MealType: status.MealType,
		}
		if existed {
			newRecord = existing
		}
		newRecord.IsParticipating = newValue
		newRecord.OverrideBy = event.UserID
		newRecord.OverrideReason = opts.Reason
		if err := store.UpsertParticipation(ctx, newRecord); err != nil {
			return nil, err
		}
		action := "CREATE"
		var oldAuditValue interface{}
		if existed {
			action = "UPDATE"
			oldAuditValue = existing
		}
		if err := writeOverrideAudit(ctx, store, event.UserID, action, "MEAL_PARTICIPATION", target.ID+"#"+date+"#"+status.MealType, "USER#"+target.ID, oldAuditValue, newRecord); err != nil {
			slog.Warn("meal override audit write failed", "targetUserID", target.ID, "date", date, "mealType", status.MealType, "error", err)
		}
		changed = append(changed, fmt.Sprintf("%s -> %s", cmdutil.DisplayMealName(status.MealType), mealOverrideDisplay(newValue)))
	}

	return &overrideResult{TargetEmail: target.Email, Entry: "meal", Date: date, Summary: strings.Join(changed, ", "), Reason: opts.Reason}, nil
}

func executeLocationOverride(ctx context.Context, store overrideStore, event payload.CommandEvent, target *repository.User, date string, opts payload.OverrideOptions) (*overrideResult, error) {
	if opts.Value != "" && opts.Value != "office" && opts.Value != "wfh" {
		return nil, userOverrideError("Location override value must be `office` or `wfh`, or omitted to toggle.")
	}

	current, err := store.GetWorkLocation(ctx, target.ID, date)
	if err != nil {
		return nil, err
	}
	newLocation := opts.Value
	if newLocation == "" {
		currentLocation := "not_set"
		if current != nil {
			currentLocation = current.Location
		}
		newLocation = toggledLocationValue(currentLocation)
	}
	newRecord := repository.WorkLocation{UserID: target.ID, Date: date, Location: newLocation, SetBy: event.UserID, Reason: opts.Reason}
	if current != nil {
		newRecord.CreatedAt = current.CreatedAt
	}
	if err := store.UpsertWorkLocation(ctx, newRecord); err != nil {
		return nil, err
	}
	action := "CREATE"
	var oldAuditValue interface{}
	if current != nil {
		action = "UPDATE"
		oldAuditValue = current
	}
	if err := writeOverrideAudit(ctx, store, event.UserID, action, "WORK_LOCATION", target.ID+"#"+date, "USER#"+target.ID, oldAuditValue, newRecord); err != nil {
		slog.Warn("location override audit write failed", "targetUserID", target.ID, "date", date, "error", err)
	}

	return &overrideResult{TargetEmail: target.Email, Entry: "location", Date: date, Summary: displayLocationLabel(newLocation), Reason: opts.Reason}, nil
}

func writeOverrideAudit(ctx context.Context, store overrideStore, actorUserID, action, entityType, entityKey, targetEntity string, oldValue, newValue interface{}) error {
	oldJSON := ""
	if action != "CREATE" && oldValue != nil {
		oldBytes, err := json.Marshal(oldValue)
		if err != nil {
			return err
		}
		oldJSON = string(oldBytes)
	}
	newJSON, err := json.Marshal(newValue)
	if err != nil {
		return err
	}
	return store.WriteAuditEntry(ctx, repository.AuditEntry{ActorUserID: actorUserID, Timestamp: time.Now().UTC(), EntityType: entityType, EntityKey: entityKey, Action: action, OldValue: oldJSON, NewValue: string(newJSON), TargetEntity: targetEntity})
}

func mealOverrideDisplay(isParticipating bool) string {
	if isParticipating {
		return "Included"
	}
	return "Opted out"
}

func formatOverrideSuccess(result *overrideResult) string {
	title := "Location override saved"
	if result.Entry == "meal" {
		title = "Meal override saved"
	}
	return fmt.Sprintf("%s for %s on %s\nUpdated: %s\nReason: %s", title, result.TargetEmail, result.Date, result.Summary, result.Reason)
}

func handleTeamSummaryCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "team_lead" && event.Role != "admin" {
		return sendWarningReply(ctx, cfg, event, "You do not have permission to use `/team-summary`.")
	}

	var opts payload.TeamSummaryOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendWarningReply(ctx, cfg, event, "Invalid command options.")
	}
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, or YYYY-MM-DD", err))
	}

	team, err := resolveTeamSummaryTeam(ctx, store, event, opts.TeamID)
	if err != nil {
		var userErr *overrideUserError
		if errors.As(err, &userErr) {
			return sendWarningReply(ctx, cfg, event, userErr.Error())
		}
		return sendErrorReply(ctx, cfg, event, "Something went wrong fetching your team. Please try again later.")
	}

	summary, err := services.GetTeamSummary(ctx, store, team.ID, date)
	if err != nil {
		return sendErrorReply(ctx, cfg, event, "Something went wrong fetching the team summary. Please try again later.")
	}

	if event.Source == "gchat" {
		rows := buildTeamSummaryRows(team, summary)
		card, _ := gchat.TeamSummaryCard(date, rows)
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendDiscordMessage(ctx, cfg, event, buildDiscordTeamSummaryMessage(team, date, summary))
}

func resolveTeamSummaryTeam(ctx context.Context, store *repository.Store, event payload.CommandEvent, requestedTeamID string) (*repository.Team, error) {
	requestedTeamID = strings.TrimSpace(requestedTeamID)
	if event.Role == "admin" {
		if requestedTeamID != "" {
			team, err := store.GetTeamByID(ctx, requestedTeamID)
			if err != nil {
				return nil, err
			}
			if team == nil {
				return nil, userOverrideError("Requested team not found.")
			}
			return team, nil
		}

		user, err := store.GetUserByID(ctx, event.UserID)
		if err != nil {
			return nil, err
		}
		if user == nil || strings.TrimSpace(user.TeamID) == "" {
			return nil, userOverrideError("Please specify `team_id` to view another team's summary.")
		}
		team, err := store.GetTeamByID(ctx, user.TeamID)
		if err != nil {
			return nil, err
		}
		if team == nil {
			return nil, userOverrideError("Requested team not found.")
		}
		return team, nil
	}

	if requestedTeamID != "" {
		return nil, userOverrideError("`team_id` can only be used by admins.")
	}
	teams, err := store.FindTeamsByLeadID(ctx, event.UserID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, userOverrideError("You are not assigned as a team lead to any team.")
	}
	team, err := store.GetTeamByID(ctx, teams[0].ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, userOverrideError("Team not found. Please try again later.")
	}
	return team, nil
}

func buildTeamSummaryRows(team *repository.Team, summary *services.TeamSummary) []gchat.TeamRow {
	rows := []gchat.TeamRow{
		{Label: "Team", Value: team.Name},
		{Label: "Members", Value: fmt.Sprintf("%d", summary.MemberCount)},
	}

	mealTypes := make([]string, 0, len(summary.MealCounts))
	for mt := range summary.MealCounts {
		mealTypes = append(mealTypes, mt)
	}
	sort.Strings(mealTypes)

	for _, mt := range mealTypes {
		rows = append(rows, gchat.TeamRow{
			Label: cmdutil.DisplayMealName(mt),
			Value: fmt.Sprintf("%d / %d", summary.MealCounts[mt], summary.MemberCount),
		})
	}

	officeCount := summary.MemberCount - summary.WFHCount
	rows = append(rows, gchat.TeamRow{
		Label: "Location",
		Value: fmt.Sprintf("Office %d / WFH %d", officeCount, summary.WFHCount),
	})

	return rows
}

func buildDiscordTeamSummaryMessage(team *repository.Team, date string, summary *services.TeamSummary) discord.Message {
	fields := []discord.EmbedField{
		{Name: "Team", Value: team.Name, Inline: true},
		{Name: "Members", Value: fmt.Sprintf("%d", summary.MemberCount), Inline: true},
		{Name: "Office / WFH", Value: fmt.Sprintf("%d / %d", summary.MemberCount-summary.WFHCount, summary.WFHCount), Inline: true},
	}

	mealTypes := make([]string, 0, len(summary.MealCounts))
	for mealType := range summary.MealCounts {
		mealTypes = append(mealTypes, mealType)
	}
	sort.Strings(mealTypes)

	for _, mealType := range mealTypes {
		fields = append(fields, discord.EmbedField{
			Name:   cmdutil.DisplayMealName(mealType),
			Value:  fmt.Sprintf("%d / %d included", summary.MealCounts[mealType], summary.MemberCount),
			Inline: true,
		})
	}

	return discord.EmbedMessage(discord.BrandEmbed("Team Summary", date, fields))
}
