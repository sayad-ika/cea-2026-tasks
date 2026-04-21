package repository

// DayStatus represents the type of day
type DayStatus string

const (
	DayStatusNormal       DayStatus = "normal"
	DayStatusOfficeClosed DayStatus = "office_closed"
	DayStatusGovtHoliday  DayStatus = "govt_holiday"
	DayStatusCelebration  DayStatus = "celebration"
	DayStatusWeekend      DayStatus = "weekend"
	DayStatusEventDay     DayStatus = "event_day"
)

// MealType represents available meal types
type MealType string

const (
	MealTypeLunch          MealType = "lunch"
	MealTypeSnacks         MealType = "snacks"
	MealTypeIftar          MealType = "iftar"
	MealTypeEventDinner    MealType = "event_dinner"
	MealTypeOptionalDinner MealType = "optional_dinner"
)

// ValidDayStatuses returns all valid day status values
func ValidDayStatuses() []string {
	return []string{
		string(DayStatusNormal),
		string(DayStatusOfficeClosed),
		string(DayStatusGovtHoliday),
		string(DayStatusCelebration),
		string(DayStatusWeekend),
		string(DayStatusEventDay),
	}
}

// ValidMealTypes returns all valid meal type values
func ValidMealTypes() []string {
	return []string{
		string(MealTypeLunch),
		string(MealTypeSnacks),
		string(MealTypeIftar),
		string(MealTypeEventDinner),
		string(MealTypeOptionalDinner),
	}
}

// IsValidDayStatus checks if a day status is valid
func IsValidDayStatus(status string) bool {
	switch DayStatus(status) {
	case DayStatusNormal, DayStatusOfficeClosed, DayStatusGovtHoliday,
		DayStatusCelebration, DayStatusWeekend, DayStatusEventDay:
		return true
	}
	return false
}

// IsValidMealType checks if a meal type is valid
func IsValidMealType(mealType string) bool {
	switch MealType(mealType) {
	case MealTypeLunch, MealTypeSnacks, MealTypeIftar,
		MealTypeEventDinner, MealTypeOptionalDinner:
		return true
	}
	return false
}

func IsValidMealOrAll(mealType string) bool {
	return mealType == "all" || IsValidMealType(mealType)
}
