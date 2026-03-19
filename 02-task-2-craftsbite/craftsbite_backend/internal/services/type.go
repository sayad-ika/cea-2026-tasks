package services

import (
	"github.com/sayad-ika/craftsbite/internal/repository"
)

type MealCount struct {
	MealType string
	OptedIn  int
	OptedOut int
}

type LocationCount struct {
	Office int
	WFH    int
}

type TeamHeadcount struct {
	TeamID         string
	TeamName       string
	MemberCount    int
	LocationCounts LocationCount
	MealCounts     map[string]MealCount
}

type HeadcountResult struct {
	Date           string
	DayStatus      string
	DayReason      string
	TotalUsers     int
	MealCounts     map[string]MealCount
	LocationCounts LocationCount
	Teams          []TeamHeadcount
}

type MemberDetail struct {
	UserID   string
	Name     string
	Meals    map[string]string
	Location string
}

type TeamSummary struct {
	MemberCount int
	MealCounts  map[string]int
	WFHCount    int
	Members     []MemberDetail
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