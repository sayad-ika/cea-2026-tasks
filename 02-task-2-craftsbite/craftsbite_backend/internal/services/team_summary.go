package services

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sayad-ika/craftsbite/internal/repository"
)


func fetchMemberData(ctx context.Context, client *dynamodb.Client, table, userID, date string) (memberData, error) {
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
		data.user, userErr = repository.GetUserByID(ctx, client, table, userID)
	}()

	go func() {
		defer wg.Done()
		data.meals, mealErr = repository.GetParticipationsByUserDate(ctx, client, table, userID, date)
	}()

	go func() {
		defer wg.Done()
		data.location, locErr = repository.GetWorkLocation(ctx, client, table, userID, date)
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

func GetTeamSummary(ctx context.Context, client *dynamodb.Client, table, teamID, date string, detail bool) (*TeamSummary, error) {
	members, err := repository.GetTeamMembers(ctx, client, table, teamID)
	if err != nil {
		return nil, fmt.Errorf("team_summary: members: %w", err)
	}
	if len(members) == 0 {
		return &TeamSummary{MealCounts: make(map[string]int)}, nil
	}

	var (
		availableMeals []string
		mealsErr       error
		memberDataList = make([]memberData, len(members))
		statuses       = make([]memberStatus, len(members))
		errs           = make([]error, len(members))
		wg             sync.WaitGroup
	)

	wg.Add(1 + len(members))

	go func() {
		defer wg.Done()
		availableMeals, mealsErr = repository.GetAvailableMeals(ctx, client, table, date)
	}()

	for i, m := range members {
		i, m := i, m
		go func() {
			defer wg.Done()
			data, err := fetchMemberData(ctx, client, table, m.UserID, date)
			if err != nil {
				errs[i] = fmt.Errorf("team_summary: member %s: %w", m.UserID, err)
				return
			}
			memberDataList[i] = data
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

	summary := &TeamSummary{
		MemberCount: len(members),
		MealCounts:  mealCounts,
		WFHCount:    wfhCount,
	}

	if detail {
		summary.Members = buildMemberDetails(memberDataList, seen)
	}

	return summary, nil
}

func buildMemberDetails(dataList []memberData, availableMeals map[string]bool) []MemberDetail {
	details := make([]MemberDetail, 0, len(dataList))
	for _, d := range dataList {
		if d.user == nil {
			continue
		}

		loc := "office"
		if d.location != nil && d.location.Location != "" {
			loc = d.location.Location
		}

		mealByType := make(map[string]bool, len(d.meals))
		for _, p := range d.meals {
			mealByType[p.MealType] = p.IsParticipating
		}

		meals := make(map[string]string, len(availableMeals))
		for mt := range availableMeals {
			if participating, ok := mealByType[mt]; !ok || participating {
				meals[mt] = "opted_in"
			} else {
				meals[mt] = "opted_out"
			}
		}

		name := d.user.Name
		if name == "" {
			name = d.user.ID
		}

		details = append(details, MemberDetail{
			UserID:   d.user.ID,
			Name:     name,
			Meals:    meals,
			Location: loc,
		})
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].Name < details[j].Name
	})

	return details
}
