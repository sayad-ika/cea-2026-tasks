package services

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
