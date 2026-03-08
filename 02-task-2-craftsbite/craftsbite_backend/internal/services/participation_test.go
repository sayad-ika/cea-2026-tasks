package services

import (
	"testing"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

func TestResolve(t *testing.T) {
	normalSchedule := &repository.DaySchedule{
		DayStatus:      "normal",
		AvailableMeals: []string{"lunch", "snacks"},
	}
	closedSchedule := &repository.DaySchedule{
		DayStatus:      "office_closed",
		AvailableMeals: []string{},
	}
	partialMeals := []string{"lunch"}

	optedOutRecord := &repository.MealParticipation{IsParticipating: false}
	optedInRecord := &repository.MealParticipation{IsParticipating: true}

	tests := []struct {
		name           string
		schedule       *repository.DaySchedule
		availableMeals []string
		record         *repository.MealParticipation
		mealType       string
		wantStatus     string
		wantSource     string
	}{
		{
			name:           "meal unavailable - schedule closed",
			schedule:       closedSchedule,
			availableMeals: []string{},
			record:         nil,
			mealType:       "lunch",
			wantStatus:     "unavailable",
			wantSource:     "day_schedule_unavailable",
		},
		{
			name:           "meal unavailable - not in available list",
			schedule:       normalSchedule,
			availableMeals: partialMeals, // only lunch
			record:         nil,
			mealType:       "snacks",
			wantStatus:     "unavailable",
			wantSource:     "day_schedule_unavailable",
		},
		{
			name:           "explicit opt-out",
			schedule:       normalSchedule,
			availableMeals: []string{"lunch", "snacks"},
			record:         optedOutRecord,
			mealType:       "lunch",
			wantStatus:     "opted_out",
			wantSource:     "explicit",
		},
		{
			name:           "explicit opt-in",
			schedule:       normalSchedule,
			availableMeals: []string{"lunch", "snacks"},
			record:         optedInRecord,
			mealType:       "lunch",
			wantStatus:     "opted_in",
			wantSource:     "explicit",
		},
		{
			name:           "no record, meal available - system default",
			schedule:       normalSchedule,
			availableMeals: []string{"lunch", "snacks"},
			record:         nil,
			mealType:       "lunch",
			wantStatus:     "opted_in",
			wantSource:     "system_default",
		},
		{
			name:           "nil schedule - all meals available, system default",
			schedule:       nil,
			availableMeals: []string{"lunch", "snacks"},
			record:         nil,
			mealType:       "lunch",
			wantStatus:     "opted_in",
			wantSource:     "system_default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.schedule, tc.availableMeals, tc.record, tc.mealType)
			if got.Status != tc.wantStatus {
				t.Errorf("Status = %q; want %q", got.Status, tc.wantStatus)
			}
			if got.Source != tc.wantSource {
				t.Errorf("Source = %q; want %q", got.Source, tc.wantSource)
			}
			if got.MealType != tc.mealType {
				t.Errorf("MealType = %q; want %q", got.MealType, tc.mealType)
			}
		})
	}
}
