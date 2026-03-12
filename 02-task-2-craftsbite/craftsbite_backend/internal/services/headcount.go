package services

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sayad-ika/craftsbite/internal/repository"
)

func GetHeadcount(ctx context.Context, client *dynamodb.Client, table, date string) (*HeadcountResult, error) {
	raw, err := fetchRawData(ctx, client, table, date)
	if err != nil {
		return nil, err
	}

	teamNames, err := fetchTeamNames(ctx, client, table, raw.users)
	if err != nil {
		return nil, err
	}

	locationByUser := buildLocationLookup(raw.locations)
	partsByUserMeal := buildParticipationLookup(raw.participations)
	availableMeals := resolveAvailableMeals(raw.schedule, raw.participations)

	teamBreakdowns := buildTeamBreakdowns(raw.users, teamNames, locationByUser, partsByUserMeal, availableMeals)
	overallMeals, overallLoc := aggregateOverallCounts(teamBreakdowns)

	return buildResult(date, raw.schedule, raw.users, overallMeals, overallLoc, teamBreakdowns), nil
}


type rawData struct {
	participations []repository.MealParticipation
	locations      []repository.WorkLocation
	users          []repository.User
	schedule       *repository.DaySchedule
}

func fetchRawData(ctx context.Context, client *dynamodb.Client, table, date string) (*rawData, error) {
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		res rawData
		err error
	)

	fetch := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := fn(); e != nil {
				mu.Lock()
				if err == nil {
					err = e
				}
				mu.Unlock()
			}
		}()
	}

	fetch(func() error {
		p, e := repository.GetParticipationsByDate(ctx, client, table, date)
		res.participations = p
		if e != nil {
			return fmt.Errorf("participations: %w", e)
		}
		return nil
	})
	fetch(func() error {
		l, e := repository.GetWorkLocationsByDate(ctx, client, table, date)
		res.locations = l
		if e != nil {
			return fmt.Errorf("locations: %w", e)
		}
		return nil
	})
	fetch(func() error {
		u, e := repository.ListActiveUsers(ctx, client, table)
		res.users = u
		if e != nil {
			return fmt.Errorf("users: %w", e)
		}
		return nil
	})
	fetch(func() error {
		s, e := repository.GetDay(ctx, client, table, date)
		res.schedule = s
		if e != nil {
			return fmt.Errorf("schedule: %w", e)
		}
		return nil
	})

	wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("headcount: %w", err)
	}
	return &res, nil
}

func fetchTeamNames(ctx context.Context, client *dynamodb.Client, table string, users []repository.User) (map[string]string, error) {
	uniqueIDs := uniqueTeamIDs(users)
	if len(uniqueIDs) == 0 {
		return map[string]string{}, nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		names   = make(map[string]string, len(uniqueIDs))
	)

	for id := range uniqueIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			t, err := repository.GetTeamByID(ctx, client, table, id)
			name := id // fallback
			if err == nil && t != nil && t.Name != "" {
				name = t.Name
			}
			mu.Lock()
			names[id] = name
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return names, nil
}

// --- Lookup Builders ---

func buildLocationLookup(locations []repository.WorkLocation) map[string]string {
	lookup := make(map[string]string, len(locations))
	for _, wl := range locations {
		lookup[wl.UserID] = wl.Location
	}
	return lookup
}

func buildParticipationLookup(participations []repository.MealParticipation) map[string]map[string]repository.MealParticipation {
	lookup := make(map[string]map[string]repository.MealParticipation)
	for _, p := range participations {
		if lookup[p.UserID] == nil {
			lookup[p.UserID] = make(map[string]repository.MealParticipation)
		}
		lookup[p.UserID][p.MealType] = p
	}
	return lookup
}

// --- Business Logic ---

func resolveAvailableMeals(schedule *repository.DaySchedule, participations []repository.MealParticipation) []string {
	if schedule != nil && len(schedule.AvailableMeals) > 0 {
		return schedule.AvailableMeals
	}
	return mealsFromParticipations(participations)
}

func mealsFromParticipations(participations []repository.MealParticipation) []string {
	seen := make(map[string]struct{})
	for _, p := range participations {
		seen[p.MealType] = struct{}{}
	}
	meals := make([]string, 0, len(seen))
	for mt := range seen {
		meals = append(meals, mt)
	}
	sort.Strings(meals)
	return meals
}

func uniqueTeamIDs(users []repository.User) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, u := range users {
		if u.TeamID != "" {
			ids[u.TeamID] = struct{}{}
		}
	}
	return ids
}

func isOptedIn(partsByUserMeal map[string]map[string]repository.MealParticipation, userID, mealType string) bool {
	p, hasRecord := partsByUserMeal[userID][mealType]
	return !hasRecord || p.IsParticipating
}

// --- Aggregation ---

func buildTeamBreakdowns(
	users []repository.User,
	teamNames map[string]string,
	locationByUser map[string]string,
	partsByUserMeal map[string]map[string]repository.MealParticipation,
	availableMeals []string,
) []TeamHeadcount {
	teamUsers := groupUsersByTeam(users)
	breakdowns := make([]TeamHeadcount, 0, len(teamUsers))

	for teamID, members := range teamUsers {
		breakdowns = append(breakdowns, buildTeamHeadcount(teamID, teamNames[teamID], members, locationByUser, partsByUserMeal, availableMeals))
	}

	sortTeamBreakdowns(breakdowns)
	return breakdowns
}

func buildTeamHeadcount(
	teamID, teamName string,
	users []repository.User,
	locationByUser map[string]string,
	partsByUserMeal map[string]map[string]repository.MealParticipation,
	availableMeals []string,
) TeamHeadcount {
	if teamID == "" {
		teamName = "Unassigned"
	}

	meals := make(map[string]MealCount)
	var loc LocationCount

	for _, u := range users {
		if locationByUser[u.ID] == "wfh" {
			loc.WFH++
		} else {
			loc.Office++
		}

		for _, mealType := range availableMeals {
			mc := meals[mealType]
			mc.MealType = mealType
			if isOptedIn(partsByUserMeal, u.ID, mealType) {
				mc.OptedIn++
			} else {
				mc.OptedOut++
			}
			meals[mealType] = mc
		}
	}

	return TeamHeadcount{
		TeamID:         teamID,
		TeamName:       teamName,
		MemberCount:    len(users),
		LocationCounts: loc,
		MealCounts:     meals,
	}
}

func aggregateOverallCounts(teams []TeamHeadcount) (map[string]MealCount, LocationCount) {
	meals := make(map[string]MealCount)
	var loc LocationCount

	for _, t := range teams {
		loc.WFH += t.LocationCounts.WFH
		loc.Office += t.LocationCounts.Office
		for mealType, mc := range t.MealCounts {
			overall := meals[mealType]
			overall.MealType = mealType
			overall.OptedIn += mc.OptedIn
			overall.OptedOut += mc.OptedOut
			meals[mealType] = overall
		}
	}

	return meals, loc
}

func groupUsersByTeam(users []repository.User) map[string][]repository.User {
	groups := make(map[string][]repository.User)
	for _, u := range users {
		groups[u.TeamID] = append(groups[u.TeamID], u)
	}
	return groups
}

func sortTeamBreakdowns(teams []TeamHeadcount) {
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].TeamID == "" {
			return false
		}
		if teams[j].TeamID == "" {
			return true
		}
		return teams[i].TeamName < teams[j].TeamName
	})
}

// --- Result Builder ---

func buildResult(date string, schedule *repository.DaySchedule, users []repository.User, meals map[string]MealCount, loc LocationCount, teams []TeamHeadcount) *HeadcountResult {
	dayStatus, dayReason := "", ""
	if schedule != nil {
		dayStatus = schedule.DayStatus
		dayReason = schedule.Reason
	}

	return &HeadcountResult{
		Date:           date,
		DayStatus:      dayStatus,
		DayReason:      dayReason,
		TotalUsers:     len(users),
		MealCounts:     meals,
		LocationCounts: loc,
		Teams:          teams,
	}
}