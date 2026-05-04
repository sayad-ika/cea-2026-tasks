package main

import (
	"context"
	"fmt"
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

func handleSelfCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	switch event.CommandName {
	case "help":
		return handleHelpCommand(ctx, deps.cfg, event)
	case "meal":
		reply := handleMeal(ctx, deps.store, deps.dateParser, deps.cutoff, event)
		if !reply.rich {
			if reply.tone == discord.NoticeToneWarning {
				return sendWarningReply(ctx, deps.cfg, event, reply.text)
			}
			if reply.tone == discord.NoticeToneError {
				return sendErrorReply(ctx, deps.cfg, event, reply.text)
			}
			return sendReply(ctx, deps.cfg, event, reply.text)
		}
		if event.Source == "gchat" {
			card, _ := gchat.MealStatusCard(reply.date, reply.changedMeals, mealStatusRows(reply.statuses))
			return sendGChatCard(ctx, deps.cfg, event, card)
		}
		return sendDiscordMessage(ctx, deps.cfg, event, buildDiscordMealMessage(reply.date, reply.statuses, reply.changedMeals))
	case "location":
		return handleLocationCommand(ctx, deps.store, deps.cfg, deps.dateParser, deps.cutoff, event)
	case "status":
		return handleStatusCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	default:
		return sendWarningReply(ctx, deps.cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleStatusCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	var opts payload.StatusOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid command options."))
	}
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}

	mealStatuses, err := services.GetUserMealStatus(ctx, store, store, event.UserID, date)
	if err != nil {
		return sendErrorReply(ctx, cfg, event, "Unable to fetch meal status. Please try again later.")
	}

	location, err := services.GetLocation(ctx, store, event.UserID, date)
	if err != nil {
		return sendErrorReply(ctx, cfg, event, "Unable to fetch location status. Please try again later.")
	}

	if event.Source == "gchat" {
		card, _ := gchat.StatusCard(date, displayLocationLabel(location.Location), mealStatusRows(mealStatuses))
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendDiscordMessage(ctx, cfg, event, buildDiscordStatusMessage(date, location.Location, mealStatuses))
}

func handleBulkLocationUpdate(ctx context.Context, store *repository.Store, cfg *appconfig.Config, cutoff *services.CutoffChecker, event payload.CommandEvent, dates []string, location string) error {
	var successDates []string
	var failedDates []string
	var errors []string

	for _, date := range dates {
		_, err := services.SetLocation(ctx, store, event.UserID, date, location, cutoff)
		if err != nil {
			failedDates = append(failedDates, date)
			errors = append(errors, fmt.Sprintf("%s: %s", date, locationErrorReply(err, date)))
		} else {
			successDates = append(successDates, date)
		}
	}

	var sb strings.Builder
	if len(successDates) > 0 {
		locLabel := "Office"
		if location == "wfh" {
			locLabel = "WFH"
		}

		fmt.Fprintf(&sb, "✓ Successfully set location to %s for %d date(s):\n", locLabel, len(successDates))
		for _, date := range successDates {
			fmt.Fprintf(&sb, "  • %s\n", date)
		}
	}

	if len(failedDates) > 0 {
		if len(successDates) > 0 {
			fmt.Fprintf(&sb, "\n")
		}
		fmt.Fprintf(&sb, "✗ Failed to update %d date(s):\n", len(failedDates))
		for _, errMsg := range errors {
			fmt.Fprintf(&sb, "  • %s\n", errMsg)
		}
	}

	return sendReply(ctx, cfg, event, sb.String())
}

func handleBulkLocationToggle(ctx context.Context, store *repository.Store, cfg *appconfig.Config, cutoff *services.CutoffChecker, event payload.CommandEvent, dates []string) error {
	var successLines []string
	var failedDates []string
	var errors []string

	for _, date := range dates {
		wl, err := toggleLocation(ctx, store, event.UserID, date, cutoff)
		if err != nil {
			failedDates = append(failedDates, date)
			errors = append(errors, fmt.Sprintf("%s: %s", date, locationErrorReply(err, date)))
		} else {
			successLines = append(successLines, fmt.Sprintf("%s -> %s", date, displayLocationLabel(wl.Location)))
		}
	}

	var sb strings.Builder
	if len(successLines) > 0 {
		fmt.Fprintf(&sb, "✓ Successfully toggled location for %d date(s):\n", len(successLines))
		for _, line := range successLines {
			fmt.Fprintf(&sb, "  • %s\n", line)
		}
	}

	if len(failedDates) > 0 {
		if len(successLines) > 0 {
			fmt.Fprintf(&sb, "\n")
		}
		fmt.Fprintf(&sb, "✗ Failed to update %d date(s):\n", len(failedDates))
		for _, errMsg := range errors {
			fmt.Fprintf(&sb, "  • %s\n", errMsg)
		}
	}

	return sendReply(ctx, cfg, event, sb.String())
}

func handleLocationCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) error {
	var opts payload.LocationOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendWarningReply(ctx, cfg, event, "Invalid command options.")
	}
	dates, err := dateParser.ParseDateRange(opts.Date)
	if err != nil {
		return sendWarningReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err))
	}

	if opts.Location != "" && opts.Location != "office" && opts.Location != "wfh" {
		return sendWarningReply(ctx, cfg, event, "Please specify location as `office` or `wfh`, or omit it to toggle.")
	}

	if len(dates) > 1 {
		if opts.Location == "" {
			return handleBulkLocationToggle(ctx, store, cfg, cutoff, event, dates)
		}
		return handleBulkLocationUpdate(ctx, store, cfg, cutoff, event, dates, opts.Location)
	}

	date := dates[0]
	var wl *repository.WorkLocation
	if opts.Location == "" {
		wl, err = toggleLocation(ctx, store, event.UserID, date, cutoff)
	} else {
		wl, err = services.SetLocation(ctx, store, event.UserID, date, opts.Location, cutoff)
	}
	if err != nil {
		return sendWarningReply(ctx, cfg, event, locationErrorReply(err, date))
	}

	mealStatuses, _ := services.GetUserMealStatus(ctx, store, store, event.UserID, date)

	if event.Source == "gchat" {
		mealText := formatMealStatusLine(mealStatuses)
		card, _ := gchat.LocationCard(displayLocationLabel(wl.Location), date, mealText)
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendDiscordMessage(ctx, cfg, event, buildDiscordLocationMessage(date, wl.Location, mealStatuses))
}

type mealReply struct {
	text         string
	rich         bool
	date         string
	statuses     []services.ResolvedStatus
	changedMeals []string
	tone         discord.NoticeTone
}

func toggleLocation(ctx context.Context, store *repository.Store, userID, date string, cutoff *services.CutoffChecker) (*repository.WorkLocation, error) {
	current, err := services.GetLocation(ctx, store, userID, date)
	if err != nil {
		return nil, err
	}
	return services.SetLocation(ctx, store, userID, date, toggledLocationValue(current.Location), cutoff)
}

type dateRangeParser interface {
	ParseDateRange(string) ([]string, error)
}

func formatMealStatusLine(statuses []services.ResolvedStatus) string {
	if len(statuses) == 0 {
		return "No meals configured"
	}
	var parts []string
	for _, s := range statuses {
		parts = append(parts, fmt.Sprintf("%s: %s", cmdutil.DisplayMealName(s.MealType), mealStatusValue(s.Status)))
	}
	return strings.Join(parts, "  ")
}

func handleBulkMealUpdate(ctx context.Context, store *repository.Store, userID string, dates []string, mealType string, isParticipating bool, cutoff *services.CutoffChecker) string {
	var successDates []string
	var failedDates []string
	var errors []string

	for _, date := range dates {
		_, err := services.UpdateParticipation(ctx, store, store, userID, date, mealType, isParticipating, cutoff)
		if err != nil {
			failedDates = append(failedDates, date)
			errors = append(errors, fmt.Sprintf("%s: %s", date, mealErrorReply(err, date)))
		} else {
			successDates = append(successDates, date)
		}
	}

	var sb strings.Builder
	if len(successDates) > 0 {
		action := "opted out"
		if isParticipating {
			action = "opted in"
		}
		mealLabel := cmdutil.DisplayMealName(mealType)

		fmt.Fprintf(&sb, "✓ Successfully %s %s for %d date(s):\n", action, mealLabel, len(successDates))
		for _, date := range successDates {
			fmt.Fprintf(&sb, "  • %s\n", date)
		}
	}

	if len(failedDates) > 0 {
		if len(successDates) > 0 {
			fmt.Fprintf(&sb, "\n")
		}
		fmt.Fprintf(&sb, "✗ Failed to update %d date(s):\n", len(failedDates))
		for _, errMsg := range errors {
			fmt.Fprintf(&sb, "  • %s\n", errMsg)
		}
	}

	return sb.String()
}

func handleBulkMealToggle(ctx context.Context, store *repository.Store, userID string, dates []string, mealType string, cutoff *services.CutoffChecker) string {
	var successLines []string
	var failedDates []string
	var errors []string

	for _, date := range dates {
		_, changedMeals, err := toggleMealStatuses(ctx, store, userID, date, mealType, cutoff)
		if err != nil {
			failedDates = append(failedDates, date)
			errors = append(errors, fmt.Sprintf("%s: %s", date, mealErrorReply(err, date)))
			continue
		}

		label := "meal status"
		if len(changedMeals) > 0 {
			label = displayMealList(changedMeals)
		}
		if mealType != "all" && mealType != "" {
			label = displayMealList([]string{mealType})
		}
		successLines = append(successLines, fmt.Sprintf("%s (%s)", date, label))
	}

	var sb strings.Builder
	if len(successLines) > 0 {
		fmt.Fprintf(&sb, "✓ Successfully toggled meal status for %d date(s):\n", len(successLines))
		for _, line := range successLines {
			fmt.Fprintf(&sb, "  • %s\n", line)
		}
	}

	if len(failedDates) > 0 {
		if len(successLines) > 0 {
			fmt.Fprintf(&sb, "\n")
		}
		fmt.Fprintf(&sb, "✗ Failed to update %d date(s):\n", len(failedDates))
		for _, errMsg := range errors {
			fmt.Fprintf(&sb, "  • %s\n", errMsg)
		}
	}

	return sb.String()
}

func handleMeal(ctx context.Context, store *repository.Store, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) mealReply {
	var opts payload.MealOptions
	if err := event.ParseOptions(&opts); err != nil {
		return mealReply{text: "Invalid command options.", tone: discord.NoticeToneWarning}
	}
	dates, err := dateParser.ParseDateRange(opts.Date)
	if err != nil {
		return mealReply{text: fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err), tone: discord.NoticeToneWarning}
	}

	if opts.Status != "" && opts.Status != "in" && opts.Status != "out" {
		return mealReply{text: "Please specify status as `in` or `out`, or omit it to toggle.", tone: discord.NoticeToneWarning}
	}
	isParticipating := opts.Status == "in"

	mealType := opts.Meal
	if mealType == "" {
		mealType = "all"
	}
	if !repository.IsValidMealOrAll(mealType) {
		return mealReply{text: fmt.Sprintf("`%s` is not a valid meal type. Choose from: lunch, snacks, iftar, event_dinner, optional_dinner, all.", mealType), tone: discord.NoticeToneWarning}
	}

	if len(dates) > 1 {
		if opts.Status == "" {
			return mealReply{text: handleBulkMealToggle(ctx, store, event.UserID, dates, mealType, cutoff)}
		}
		return mealReply{text: handleBulkMealUpdate(ctx, store, event.UserID, dates, mealType, isParticipating, cutoff)}
	}

	date := dates[0]
	var (
		statuses     []services.ResolvedStatus
		changedMeals []string
	)
	if opts.Status == "" {
		statuses, changedMeals, err = toggleMealStatuses(ctx, store, event.UserID, date, mealType, cutoff)
	} else {
		statuses, err = services.UpdateParticipation(ctx, store, store, event.UserID, date, mealType, isParticipating, cutoff)
	}
	if err != nil {
		return mealReply{text: mealErrorReply(err, date), tone: discord.NoticeToneWarning}
	}

	if opts.Status != "" {
		if mealType == "all" {
			for _, s := range statuses {
				if s.Status != "unavailable" {
					changedMeals = append(changedMeals, s.MealType)
				}
			}
		} else {
			changedMeals = []string{mealType}
		}
	}

	return mealReply{
		rich:         true,
		date:         date,
		statuses:     statuses,
		changedMeals: changedMeals,
		text:         formatMealStatusWithChanges(date, statuses, changedMeals),
	}
}

func toggleMealStatuses(ctx context.Context, store *repository.Store, userID, date, mealType string, cutoff *services.CutoffChecker) ([]services.ResolvedStatus, []string, error) {
	targets, err := mealToggleTargets(ctx, store, userID, date, mealType)
	if err != nil {
		return nil, nil, err
	}

	var statuses []services.ResolvedStatus
	changedMeals := make([]string, 0, len(targets))
	for _, target := range targets {
		statuses, err = services.UpdateParticipation(ctx, store, store, userID, date, target.MealType, target.Status != "opted_in", cutoff)
		if err != nil {
			return nil, nil, err
		}
		changedMeals = append(changedMeals, target.MealType)
	}

	return statuses, changedMeals, nil
}

func mealToggleTargets(ctx context.Context, store *repository.Store, userID, date, mealType string) ([]services.ResolvedStatus, error) {
	schedule, err := store.GetDay(ctx, date)
	if err != nil {
		return nil, err
	}
	if schedule != nil && (schedule.DayStatus == "office_closed" || schedule.DayStatus == "govt_holiday") {
		return nil, services.ErrDayClosed
	}

	availableMeals, err := store.GetAvailableMeals(ctx, date)
	if err != nil {
		return nil, err
	}
	if len(availableMeals) == 0 {
		return nil, services.ErrNoMeals
	}

	statuses, err := services.GetUserMealStatus(ctx, store, store, userID, date)
	if err != nil {
		return nil, err
	}
	if mealType == "" || mealType == "all" {
		return statuses, nil
	}
	for _, status := range statuses {
		if status.MealType == mealType {
			return []services.ResolvedStatus{status}, nil
		}
	}
	return nil, services.ErrMealUnavailable
}

func mealErrorReply(err error, date string) string {
	switch err {
	case services.ErrPastDate:
		return "Cannot update participation for a past date."
	case services.ErrCutoffPassed:
		return fmt.Sprintf("Updates for %s are closed. Cutoff was %s at 9:00 PM.", date, prevDay(date))
	case services.ErrTooFarAhead:
		return fmt.Sprintf("Cannot update participation for %s — that's more than 7 days away.", date)
	case services.ErrDayClosed:
		return fmt.Sprintf("Office is closed on %s — no meals are available.", date)
	case services.ErrNoMeals:
		return fmt.Sprintf("No meals are configured for %s.", date)
	case services.ErrMealUnavailable:
		return fmt.Sprintf("That meal is not available on %s.", date)
	default:
		return "Something went wrong. Please try again later."
	}
}

func locationErrorReply(err error, date string) string {
	switch err {
	case services.ErrLocationPastDate:
		return "Cannot set work location for a past date."
	case services.ErrLocationCutoffPassed:
		return fmt.Sprintf("Updates for %s are closed. Cutoff was %s at 9:00 PM.", date, prevDay(date))
	case services.ErrLocationTooFarAhead:
		return fmt.Sprintf("Cannot set work location for %s — that's more than 7 days away.", date)
	default:
		return "Something went wrong. Please try again later."
	}
}

func formatLocationStatus(date, location string, mealStatuses []services.ResolvedStatus) string {
	locIcon := "🏢"
	locLabel := "Office"
	if location == "wfh" {
		locIcon = "🏠"
		locLabel = "WFH"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Updated! Status for %s:\n → %s %s", date, locIcon, locLabel)

	for _, s := range mealStatuses {
		icon := "✗"
		if s.Status == "opted_in" {
			icon = "✓"
		} else if s.Status == "unavailable" {
			icon = "—"
		}
		fmt.Fprintf(&sb, "  %s %s", cmdutil.DisplayMealName(s.MealType), icon)
	}
	return sb.String()
}

func prevDay(targetDate string) string {
	t, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return targetDate + " (previous day)"
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
}

func formatMealStatusWithChanges(date string, statuses []services.ResolvedStatus, changedMeals []string) string {
	if len(statuses) == 0 {
		return fmt.Sprintf("No meals are available on %s.", date)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Updated! Meal status for %s:", date)
	for _, s := range statuses {
		icon := "✗"
		if s.Status == "opted_in" {
			icon = "✓"
		} else if s.Status == "unavailable" {
			icon = "—"
		}

		prefix := "  "
		if containsString(changedMeals, s.MealType) {
			prefix = " →"
		}

		fmt.Fprintf(&sb, "%s %s %s", prefix, cmdutil.DisplayMealName(s.MealType), icon)
	}
	return sb.String()
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func mealStatusRows(statuses []services.ResolvedStatus) []gchat.TeamRow {
	rows := make([]gchat.TeamRow, 0, len(statuses))
	for _, status := range statuses {
		rows = append(rows, gchat.TeamRow{
			Label: cmdutil.DisplayMealName(status.MealType),
			Value: mealStatusValue(status.Status),
		})
	}
	return rows
}

func buildDiscordStatusMessage(date, location string, mealStatuses []services.ResolvedStatus) discord.Message {
	fields := []discord.EmbedField{{Name: "Location", Value: displayLocationLabel(location), Inline: true}}
	if len(mealStatuses) == 0 {
		fields = append(fields, discord.EmbedField{Name: "Meals", Value: "No meals configured"})
	} else {
		for _, status := range mealStatuses {
			fields = append(fields, discord.EmbedField{
				Name:   cmdutil.DisplayMealName(status.MealType),
				Value:  mealStatusValue(status.Status),
				Inline: true,
			})
		}
	}
	return discord.EmbedMessage(discord.BrandEmbed("Status Snapshot", date, fields))
}

func buildDiscordLocationMessage(date, location string, mealStatuses []services.ResolvedStatus) discord.Message {
	fields := []discord.EmbedField{{Name: "Work location", Value: displayLocationLabel(location), Inline: true}}
	if len(mealStatuses) == 0 {
		fields = append(fields, discord.EmbedField{Name: "Meals", Value: "No meals configured"})
	} else {
		for _, status := range mealStatuses {
			fields = append(fields, discord.EmbedField{
				Name:   cmdutil.DisplayMealName(status.MealType),
				Value:  mealStatusValue(status.Status),
				Inline: true,
			})
		}
	}
	return discord.EmbedMessage(discord.BrandEmbed("Location Updated", date, fields))
}

func buildDiscordMealMessage(date string, statuses []services.ResolvedStatus, changedMeals []string) discord.Message {
	fields := []discord.EmbedField{}
	if len(changedMeals) > 0 {
		fields = append(fields, discord.EmbedField{Name: "Updated meals", Value: displayMealList(changedMeals)})
	}
	for _, status := range statuses {
		fields = append(fields, discord.EmbedField{
			Name:   cmdutil.DisplayMealName(status.MealType),
			Value:  mealStatusValue(status.Status),
			Inline: true,
		})
	}
	return discord.EmbedMessage(discord.BrandEmbed("Meal Status Updated", date, fields))
}

func mealStatusValue(status string) string {
	switch status {
	case "opted_in":
		return "Included"
	case "unavailable":
		return "Unavailable"
	default:
		return "Opted out"
	}
}

func displayLocationLabel(location string) string {
	switch location {
	case "wfh":
		return "WFH"
	case "not_set", "":
		return "Not set"
	default:
		return "Office"
	}
}

func toggledLocationValue(location string) string {
	if location == "office" {
		return "wfh"
	}
	return "office"
}

func displayMealList(meals []string) string {
	labels := make([]string, 0, len(meals))
	for _, meal := range meals {
		labels = append(labels, cmdutil.DisplayMealName(meal))
	}
	return strings.Join(labels, ", ")
}

func formatStatusView(date, location string, mealStatuses []services.ResolvedStatus) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Status for %s:\n", date)

	locIcon := "🏢"
	locLabel := "Office"
	if location == "wfh" {
		locIcon = "🏠"
		locLabel = "WFH"
	} else if location == "not_set" {
		locIcon = "❓"
		locLabel = "Not Set"
	}
	fmt.Fprintf(&sb, "  %s %s", locIcon, locLabel)

	if len(mealStatuses) == 0 {
		fmt.Fprintf(&sb, "\n  No meals configured")
	} else {
		for _, s := range mealStatuses {
			icon := "✗"
			if s.Status == "opted_in" {
				icon = "✓"
			} else if s.Status == "unavailable" {
				icon = "—"
			}
			fmt.Fprintf(&sb, "  %s %s", cmdutil.DisplayMealName(s.MealType), icon)
		}
	}

	return sb.String()
}
