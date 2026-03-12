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

type WorkLocation struct {
	UserID    string
	Date      string
	Location  string
	SetBy     string
	Reason    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type workLocationItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	GSI1PK    string `dynamodbav:"GSI1PK"`
	GSI1SK    string `dynamodbav:"GSI1SK"`
	UserID    string `dynamodbav:"user_id"`
	Date      string `dynamodbav:"date"`
	Location  string `dynamodbav:"location"`
	SetBy     string `dynamodbav:"set_by"`
	Reason    string `dynamodbav:"reason"`
	WFHMonth  string `dynamodbav:"wfh_month,omitempty"`
	CreatedAt string `dynamodbav:"created_at"`
	UpdatedAt string `dynamodbav:"updated_at"`
}

type User struct {
	ID     string
	Email  string
	Name   string
	Role   string
	TeamID string
	Active bool
}

type userProfileItem struct {
	ID     string `dynamodbav:"id"`
	Email  string `dynamodbav:"email"`
	Name   string `dynamodbav:"name"`
	Role   string `dynamodbav:"role"`
	TeamID string `dynamodbav:"team_id"`
	Active bool   `dynamodbav:"active"`
}

type Team struct {
	ID          string
	Name        string
	Description string
	LeadID      string
}

type teamMetadataItem struct {
	ID          string `dynamodbav:"id"`
	Name        string `dynamodbav:"name"`
	Description string `dynamodbav:"description"`
	TeamLeadID  string `dynamodbav:"team_lead_id"`
}

