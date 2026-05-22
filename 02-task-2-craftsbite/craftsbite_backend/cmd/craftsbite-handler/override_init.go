package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/dateutil"
	"github.com/sayad-ika/craftsbite/internal/payload"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

const (
	overrideInitActionOpen        = "open"
	overrideInitActionTeamChange  = "team_change"
	overrideInitActionEdit        = "edit"
	overrideInitActionDetailsEdit = "details_edit"
	overrideInitActionDetailsSave = "details_save"
	overrideInitActionApply       = "apply"
	overrideInitActionCancel      = "cancel"

	overrideInitTargetWholeTeam = "whole_team"
	overrideInitTargetMembers   = "members"

	overrideInitIndexRefPrefix = "idx:"
)

type overrideInitStore interface {
	overrideStore
	GetUserByID(ctx context.Context, userID string) (*repository.User, error)
	ListActiveTeams(ctx context.Context) ([]repository.Team, error)
}

type overrideInitDraft struct {
	Action     string
	Date       string
	TeamID     string
	TargetMode string
	Members    []string
	Entry      string
	Meal       string
	Value      string
	Reason     string
}

type overrideInitState struct {
	Draft   overrideInitDraft
	Teams   []repository.Team
	Team    *repository.Team
	Members []repository.User
}

type overrideInitBulkResult struct {
	Date       string
	TeamName   string
	Entry      string
	TargetMode string
	Requested  int
	Succeeded  int
	Failed     int
	Summary    string
	Reason     string
	Failures   []string
}

func handleOverrideInitCommand(ctx context.Context, store overrideInitStore, cfg *appconfig.Config, dateParser *dateutil.DateParser, event payload.CommandEvent) error {
	if event.Role != "team_lead" && event.Role != "admin" {
		return sendWarningReply(ctx, cfg, event, "You do not have permission to use `/override-init`.")
	}
	if event.Source == "discord" {
		return handleDiscordOverrideInitInteraction(ctx, cfg, store, dateParser, event)
	}
	if event.Source == "gchat" {
		return handleGChatOverrideInitInteraction(ctx, cfg, store, dateParser, event)
	}
	return sendWarningReply(ctx, cfg, event, "`/override-init` is currently available in Google Chat and Discord.")
}

func overrideInitDraftFromOptions(opts payload.OverrideInitOptions) overrideInitDraft {
	action := strings.TrimSpace(opts.Action)
	if action == "" {
		action = overrideInitActionOpen
	}
	return overrideInitDraft{
		Action:     action,
		Date:       strings.TrimSpace(opts.Date),
		TeamID:     strings.TrimSpace(opts.TeamID),
		TargetMode: strings.ToLower(strings.TrimSpace(opts.TargetMode)),
		Members:    normalizeOverrideInitStrings(opts.Members),
		Entry:      strings.ToLower(strings.TrimSpace(opts.Entry)),
		Meal:       strings.ToLower(strings.TrimSpace(opts.Meal)),
		Value:      strings.ToLower(strings.TrimSpace(opts.Value)),
		Reason:     strings.TrimSpace(opts.Reason),
	}
}

func overrideInitOptionsFromDraft(draft overrideInitDraft) payload.OverrideInitOptions {
	return payload.OverrideInitOptions{
		Action:     draft.Action,
		Date:       draft.Date,
		TeamID:     draft.TeamID,
		TargetMode: draft.TargetMode,
		Members:    append([]string(nil), draft.Members...),
		Entry:      draft.Entry,
		Meal:       draft.Meal,
		Value:      draft.Value,
		Reason:     draft.Reason,
	}
}

func normalizeOverrideInitStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func loadOverrideInitState(ctx context.Context, store overrideInitStore, event payload.CommandEvent, draft overrideInitDraft) (overrideInitState, error) {
	teams, err := loadOverrideInitScopedTeams(ctx, store, event)
	if err != nil {
		return overrideInitState{}, err
	}
	state := overrideInitState{Draft: draft, Teams: teams}
	if len(teams) == 0 {
		return state, nil
	}

	team := resolveOverrideInitTeamRef(teams, draft.TeamID)
	if team == nil && draft.TeamID == "" && len(teams) == 1 {
		team = &teams[0]
		state.Draft.TeamID = team.ID
	}
	if team == nil {
		return state, nil
	}

	members, err := loadOverrideInitTeamMembers(ctx, store, team.ID)
	if err != nil {
		return overrideInitState{}, err
	}
	state.Team = team
	state.Members = members
	return state, nil
}

func loadOverrideInitScopedTeams(ctx context.Context, store overrideInitStore, event payload.CommandEvent) ([]repository.Team, error) {
	var teams []repository.Team
	var err error
	if event.Role == "admin" {
		teams, err = store.ListActiveTeams(ctx)
	} else {
		teams, err = store.FindTeamsByLeadID(ctx, event.UserID)
	}
	if err != nil {
		return nil, err
	}

	activeTeams := make([]repository.Team, 0, len(teams))
	for _, team := range teams {
		if team.Active {
			activeTeams = append(activeTeams, team)
		}
	}
	sort.Slice(activeTeams, func(i, j int) bool {
		if activeTeams[i].Name == activeTeams[j].Name {
			return activeTeams[i].ID < activeTeams[j].ID
		}
		return activeTeams[i].Name < activeTeams[j].Name
	})
	return activeTeams, nil
}

func loadOverrideInitTeamMembers(ctx context.Context, store overrideInitStore, teamID string) ([]repository.User, error) {
	memberships, err := store.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	members := make([]repository.User, 0, len(memberships))
	for _, membership := range memberships {
		user, err := store.GetUserByID(ctx, membership.UserID)
		if err != nil {
			return nil, err
		}
		if user == nil || !user.Active {
			continue
		}
		members = append(members, *user)
	}
	sort.Slice(members, func(i, j int) bool {
		left := overrideInitUserLabel(members[i])
		right := overrideInitUserLabel(members[j])
		if left == right {
			return members[i].ID < members[j].ID
		}
		return left < right
	})
	return members, nil
}

func resolveOverrideInitTeamRef(teams []repository.Team, ref string) *repository.Team {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	if index, ok := parseOverrideInitIndexRef(ref); ok {
		if index >= 0 && index < len(teams) {
			return &teams[index]
		}
		return nil
	}
	for i := range teams {
		if teams[i].ID == ref {
			return &teams[i]
		}
	}
	return nil
}

func resolveOverrideInitMemberRef(members []repository.User, ref string) *repository.User {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	if index, ok := parseOverrideInitIndexRef(ref); ok {
		if index >= 0 && index < len(members) {
			return &members[index]
		}
		return nil
	}
	for i := range members {
		if members[i].ID == ref {
			return &members[i]
		}
	}
	return nil
}

func overrideInitIndexRef(index int) string {
	return overrideInitIndexRefPrefix + strconv.Itoa(index)
}

func parseOverrideInitIndexRef(ref string) (int, bool) {
	if !strings.HasPrefix(ref, overrideInitIndexRefPrefix) {
		return 0, false
	}
	index, err := strconv.Atoi(strings.TrimPrefix(ref, overrideInitIndexRefPrefix))
	if err != nil {
		return 0, false
	}
	return index, true
}

func selectedOverrideInitTargets(state overrideInitState) ([]repository.User, error) {
	if state.Team == nil {
		return nil, userOverrideError("Choose a team before saving the override.")
	}
	if len(state.Members) == 0 {
		return nil, userOverrideError("The selected team has no active members to override.")
	}

	switch state.Draft.TargetMode {
	case overrideInitTargetWholeTeam:
		return append([]repository.User(nil), state.Members...), nil
	case overrideInitTargetMembers:
		if len(state.Draft.Members) == 0 {
			return nil, userOverrideError("Choose at least one team member before saving the override.")
		}
		selected := make([]repository.User, 0, len(state.Draft.Members))
		seen := make(map[string]struct{}, len(state.Draft.Members))
		for _, ref := range state.Draft.Members {
			member := resolveOverrideInitMemberRef(state.Members, ref)
			if member == nil {
				return nil, userOverrideError("One or more selected members are no longer available for this team.")
			}
			if _, ok := seen[member.ID]; ok {
				continue
			}
			seen[member.ID] = struct{}{}
			selected = append(selected, *member)
		}
		return selected, nil
	default:
		return nil, userOverrideError("Choose whether to override the whole team or specific members.")
	}
}

func applyOverrideInitDraft(ctx context.Context, store overrideInitStore, dateParser *dateutil.DateParser, event payload.CommandEvent, draft overrideInitDraft) (*overrideInitBulkResult, error) {
	state, err := loadOverrideInitState(ctx, store, event, draft)
	if err != nil {
		return nil, err
	}
	if state.Team == nil && strings.TrimSpace(draft.TeamID) != "" {
		return nil, userOverrideError("Selected team is not available to you.")
	}
	targets, err := selectedOverrideInitTargets(state)
	if err != nil {
		return nil, err
	}

	date, err := dateParser.ParseDateWithDefaults(draft.Date)
	if err != nil {
		return nil, userOverrideError(fmt.Sprintf("Invalid date: %v\nUse: tomorrow (default), today, +N, or YYYY-MM-DD", err))
	}
	if draft.Entry != "meal" && draft.Entry != "location" {
		return nil, userOverrideError("Choose whether to override meal or location.")
	}
	if draft.Reason == "" {
		return nil, userOverrideError("Override reason is required.")
	}
	if draft.Entry == "meal" {
		if draft.Meal == "" {
			draft.Meal = "all"
		}
		if !repository.IsValidMealOrAll(draft.Meal) {
			return nil, userOverrideError("Choose a valid meal type before saving.")
		}
		if draft.Value != "in" && draft.Value != "out" {
			return nil, userOverrideError("Choose whether the meal should be included or opted out.")
		}
	} else if draft.Value != "office" && draft.Value != "wfh" {
		return nil, userOverrideError("Choose whether the location should be Office or WFH.")
	}

	result := &overrideInitBulkResult{
		Date:       date,
		TeamName:   state.Team.Name,
		Entry:      draft.Entry,
		TargetMode: draft.TargetMode,
		Requested:  len(targets),
		Reason:     draft.Reason,
	}

	for _, target := range targets {
		target := target
		overrideOpts := payload.OverrideOptions{
			Target: target.Email,
			Entry:  draft.Entry,
			Date:   date,
			Meal:   draft.Meal,
			Value:  draft.Value,
			Reason: draft.Reason,
		}
		var single *overrideResult
		var applyErr error
		if draft.Entry == "meal" {
			single, applyErr = executeMealOverride(ctx, store, event, &target, date, overrideOpts)
		} else {
			single, applyErr = executeLocationOverride(ctx, store, event, &target, date, overrideOpts)
		}
		if applyErr != nil {
			result.Failed++
			result.Failures = append(result.Failures, fmt.Sprintf("%s: %s", target.Email, overrideInitFailureMessage(applyErr)))
			continue
		}
		result.Succeeded++
		if result.Summary == "" && single != nil {
			result.Summary = single.Summary
		}
	}
	return result, nil
}

func overrideInitFailureMessage(err error) string {
	var userErr *overrideUserError
	if errors.As(err, &userErr) {
		return userErr.Error()
	}
	return "save failed"
}

func overrideInitResultTitle(result *overrideInitBulkResult) string {
	if result == nil || result.Succeeded == 0 {
		return "Override not saved"
	}
	if result.Failed > 0 {
		return "Override partially saved"
	}
	if result.Requested == 1 {
		return "Override saved"
	}
	return "Overrides saved"
}

func overrideInitResultLines(result *overrideInitBulkResult) []string {
	if result == nil {
		return []string{"No override result is available."}
	}
	entryLabel := "Location"
	if result.Entry == "meal" {
		entryLabel = "Meal"
	}
	lines := []string{
		fmt.Sprintf("%s override for %s on %s", entryLabel, result.TeamName, result.Date),
		fmt.Sprintf("Saved: %d of %d", result.Succeeded, result.Requested),
	}
	if result.Summary != "" {
		lines = append(lines, "Updated: "+result.Summary)
	}
	if result.Reason != "" {
		lines = append(lines, "Reason: "+result.Reason)
	}
	if result.Failed > 0 {
		lines = append(lines, fmt.Sprintf("Failed: %d", result.Failed))
		limit := len(result.Failures)
		if limit > 5 {
			limit = 5
		}
		lines = append(lines, result.Failures[:limit]...)
		if len(result.Failures) > limit {
			lines = append(lines, fmt.Sprintf("...and %d more", len(result.Failures)-limit))
		}
	}
	return lines
}

func overrideInitUserLabel(user repository.User) string {
	if strings.TrimSpace(user.Name) != "" {
		return user.Name
	}
	if strings.TrimSpace(user.Email) != "" {
		return user.Email
	}
	return user.ID
}

func overrideInitTeamLabel(team repository.Team) string {
	if strings.TrimSpace(team.Name) != "" {
		return team.Name
	}
	return team.ID
}
