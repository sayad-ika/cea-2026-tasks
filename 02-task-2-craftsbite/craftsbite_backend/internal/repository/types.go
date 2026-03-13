package repository

import "time"

type MealParticipation struct {
	UserID          string
	Date            string
	MealType        string
	IsParticipating bool
	OverrideBy      string
	OverrideReason  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DaySchedule struct {
	Date           string
	DayStatus      string
	AvailableMeals []string
	Reason         string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type mealItem struct {
	PK              string `dynamodbav:"PK"`
	SK              string `dynamodbav:"SK"`
	GSI1PK          string `dynamodbav:"GSI1PK"`
	GSI1SK          string `dynamodbav:"GSI1SK"`
	UserID          string `dynamodbav:"user_id"`
	Date            string `dynamodbav:"date"`
	MealType        string `dynamodbav:"meal_type"`
	IsParticipating bool   `dynamodbav:"is_participating"`
	OverrideBy      string `dynamodbav:"override_by"`
	OverrideReason  string `dynamodbav:"override_reason"`
	CreatedAt       string `dynamodbav:"created_at"`
	UpdatedAt       string `dynamodbav:"updated_at"`
}

type dayScheduleItem struct {
	PK             string   `dynamodbav:"PK"`
	SK             string   `dynamodbav:"SK"`
	Date           string   `dynamodbav:"date"`
	DayStatus      string   `dynamodbav:"day_status"`
	AvailableMeals []string `dynamodbav:"available_meals"`
	Reason         string   `dynamodbav:"reason"`
	CreatedBy      string   `dynamodbav:"created_by"`
	CreatedAt      string   `dynamodbav:"created_at"`
	UpdatedAt      string   `dynamodbav:"updated_at"`
}

type dayMealsItem struct {
	PK    string   `dynamodbav:"PK"`
	SK    string   `dynamodbav:"SK"`
	Meals []string `dynamodbav:"meals"`
}
