package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sayad-ika/craftsbite/internal/cmdutil"
	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func handleSelfCommand(ctx context.Context, deps handlerDeps, event payload.CommandEvent) error {
	switch event.CommandName {
	case "meal":
		replyContent := handleMeal(ctx, deps.store, deps.dateParser, deps.cutoff, event)
		return sendReply(ctx, deps.cfg, event, replyContent)
	case "location":
		return handleLocationCommand(ctx, deps.store, deps.cfg, deps.dateParser, deps.cutoff, event)
	case "status":
		return handleStatusCommand(ctx, deps.store, deps.cfg, deps.dateParser, event)
	default:
		return sendReply(ctx, deps.cfg, event, fmt.Sprintf("Unknown command: /%s", event.CommandName))
	}
}

func handleStatusCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	var opts payload.StatusOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Invalid command options."))
	}
	date, err := dateParser.ParseDateWithDefaults(opts.Date)
	if err != nil {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}

	mealStatuses, err := services.GetUserMealStatus(ctx, store, store, event.UserID, date)
	if err != nil {
		return sendReply(ctx, cfg, event, "Unable to fetch meal status. Please try again later.")
	}

	location, err := services.GetLocation(ctx, store, event.UserID, date)
	if err != nil {
		return sendReply(ctx, cfg, event, "Unable to fetch location status. Please try again later.")
	}

	return sendReply(ctx, cfg, event, formatStatusView(date, location.Location, mealStatuses))
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

func handleLocationCommand(ctx context.Context, store *repository.Store, cfg *appconfig.Config, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) error {
	var opts payload.LocationOptions
	if err := event.ParseOptions(&opts); err != nil {
		return sendReply(ctx, cfg, event, "Invalid command options.")
	}
	dates, err := dateParser.ParseDateRange(opts.Date)
	if err != nil {
		return sendReply(ctx, cfg, event, fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err))
	}

	if opts.Location != "office" && opts.Location != "wfh" {
		return sendReply(ctx, cfg, event, "Please specify location as `office` or `wfh`.")
	}

	if len(dates) > 1 {
		return handleBulkLocationUpdate(ctx, store, cfg, cutoff, event, dates, opts.Location)
	}

	date := dates[0]
	wl, err := services.SetLocation(ctx, store, event.UserID, date, opts.Location, cutoff)
	if err != nil {
		return sendReply(ctx, cfg, event, locationErrorReply(err, date))
	}

	mealStatuses, _ := services.GetUserMealStatus(ctx, store, store, event.UserID, date)

	if event.Source == "gchat" {
		mealText := formatMealStatusLine(mealStatuses)
		card, _ := gchat.LocationCard(wl.Location, date, mealText)
		return sendGChatCard(ctx, cfg, event, card)
	}

	return sendReply(ctx, cfg, event, formatLocationStatus(date, wl.Location, mealStatuses))
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
		icon := "✗"
		if s.Status == "opted_in" {
			icon = "✓"
		} else if s.Status == "unavailable" {
			icon = "—"
		}
		parts = append(parts, fmt.Sprintf("%s %s", cmdutil.DisplayMealName(s.MealType), icon))
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

func handleMeal(ctx context.Context, store *repository.Store, dateParser *dateutil.DateParser, cutoff *services.CutoffChecker, event payload.CommandEvent) string {
	var opts payload.MealOptions
	if err := event.ParseOptions(&opts); err != nil {
		return "Invalid command options."
	}
	dates, err := dateParser.ParseDateRange(opts.Date)
	if err != nil {
		return fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), +N, YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week", err)
	}

	if opts.Status != "in" && opts.Status != "out" {
		return "Please specify status as `in` or `out`."
	}
	isParticipating := opts.Status == "in"

	mealType := opts.Meal
	if mealType == "" {
		mealType = "all"
	}
	if !repository.IsValidMealOrAll(mealType) {
		return fmt.Sprintf("`%s` is not a valid meal type. Choose from: lunch, snacks, iftar, event_dinner, optional_dinner, all.", mealType)
	}

	if len(dates) > 1 {
		return handleBulkMealUpdate(ctx, store, event.UserID, dates, mealType, isParticipating, cutoff)
	}

	date := dates[0]
	statuses, err := services.UpdateParticipation(ctx, store, store, event.UserID, date, mealType, isParticipating, cutoff)
	if err != nil {
		return mealErrorReply(err, date)
	}

	var changedMeals []string
	if mealType == "all" {
		for _, s := range statuses {
			if s.Status != "unavailable" {
				changedMeals = append(changedMeals, s.MealType)
			}
		}
	} else {
		changedMeals = []string{mealType}
	}

	return formatMealStatusWithChanges(date, statuses, changedMeals)
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
