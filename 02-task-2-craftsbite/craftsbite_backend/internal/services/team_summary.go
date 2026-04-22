package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

type TeamSummary struct {
	MemberCount int
	MealCounts  map[string]int
	WFHCount    int
}

type memberStatus struct {
	meals    []repository.MealParticipation
	location string
}

type memberData struct {
	user     *repository.User
	meals    []repository.MealParticipation
	location *repository.WorkLocation
}

func fetchMemberData(ctx context.Context, store TeamSummaryStore, userID, date string) (memberData, error) {
	var (
		data    memberData
		userErr error
		mealErr error
		locErr  error
		wg      sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		data.user, userErr = store.GetUserByID(ctx, userID)
	}()

	go func() {
		defer wg.Done()
		data.meals, mealErr = store.GetParticipationsByUserDate(ctx, userID, date)
	}()

	go func() {
		defer wg.Done()
		data.location, locErr = store.GetWorkLocation(ctx, userID, date)
	}()

	wg.Wait()

	if userErr != nil {
		return memberData{}, fmt.Errorf("user: %w", userErr)
	}
	if mealErr != nil {
		return memberData{}, fmt.Errorf("meals: %w", mealErr)
	}
	if locErr != nil {
		return memberData{}, fmt.Errorf("location: %w", locErr)
	}

	return data, nil
}

func buildMemberStatus(d memberData) memberStatus {
	loc := "office"
	if d.location != nil && d.location.Location != "" {
		loc = d.location.Location
	}
	return memberStatus{
		meals:    d.meals,
		location: loc,
	}
}

func GetTeamSummary(ctx context.Context, store TeamSummaryStore, teamID, date string) (*TeamSummary, error) {
	members, err := store.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("team_summary: members: %w", err)
	}
	if len(members) == 0 {
		return &TeamSummary{MealCounts: make(map[string]int)}, nil
	}

	var (
		availableMeals []string
		mealsErr       error
		statuses       = make([]memberStatus, len(members))
		errs           = make([]error, len(members))
		wg             sync.WaitGroup
	)

	wg.Add(1 + len(members))

	go func() {
		defer wg.Done()
		availableMeals, mealsErr = store.GetAvailableMeals(ctx, date)
	}()

	for i, m := range members {
		i, m := i, m
		go func() {
			defer wg.Done()
			data, err := fetchMemberData(ctx, store, m.UserID, date)
			if err != nil {
				errs[i] = fmt.Errorf("team_summary: member %s: %w", m.UserID, err)
				return
			}
			statuses[i] = buildMemberStatus(data)
		}()
	}

	wg.Wait()

	if mealsErr != nil {
		return nil, fmt.Errorf("team_summary: available meals: %w", mealsErr)
	}
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}

	seen := make(map[string]bool, len(availableMeals))
	for _, mt := range availableMeals {
		seen[mt] = true
	}
	if len(seen) == 0 {
		for _, s := range statuses {
			for _, p := range s.meals {
				seen[p.MealType] = true
			}
		}
	}

	mealCounts := make(map[string]int, len(seen))
	wfhCount := 0

	for _, s := range statuses {
		if s.location == "wfh" {
			wfhCount++
		}

		mealByType := make(map[string]bool, len(s.meals))
		for _, p := range s.meals {
			mealByType[p.MealType] = p.IsParticipating
		}

		for mt := range seen {
			if participating, ok := mealByType[mt]; !ok || participating {
				mealCounts[mt]++
			}
		}
	}

	return &TeamSummary{
		MemberCount: len(members),
		MealCounts:  mealCounts,
		WFHCount:    wfhCount,
	}, nil
}
