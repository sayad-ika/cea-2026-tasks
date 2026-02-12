package models

// MealType represents the type of meal
type MealType string

const (
	MealTypeLunch          MealType = "lunch"
	MealTypeSnacks         MealType = "snacks"
	MealTypeIftar          MealType = "iftar"
	MealTypeEventDinner    MealType = "event_dinner"
	MealTypeOptionalDinner MealType = "optional_dinner"
)

// IsValid checks if the meal type is valid
func (m MealType) IsValid() bool {
	switch m {
	case MealTypeLunch, MealTypeSnacks, MealTypeIftar, MealTypeEventDinner, MealTypeOptionalDinner:
		return true
	}
	return false
}

// String returns the string representation of the meal type
func (m MealType) String() string {
	return string(m)
}

// WorkLocation represents where a user will work on a specific date
type WorkLocation string

const (
	WorkLocationOffice WorkLocation = "office"
	WorkLocationWFH    WorkLocation = "wfh"
)

// IsValid checks if the work location is valid
func (w WorkLocation) IsValid() bool {
	switch w {
	case WorkLocationOffice, WorkLocationWFH:
		return true
	}
	return false
}

// String returns the string representation of the work location
func (w WorkLocation) String() string {
	return string(w)
}

// WorkLocationHistoryAction represents an action in work location history
type WorkLocationHistoryAction string

const (
	WorkLocationHistoryActionSelfSet             WorkLocationHistoryAction = "self_set"
	WorkLocationHistoryActionLeadOverride        WorkLocationHistoryAction = "lead_override"
	WorkLocationHistoryActionAdminOverride       WorkLocationHistoryAction = "admin_override"
	WorkLocationHistoryActionGlobalPolicyCreated WorkLocationHistoryAction = "global_policy_created"
	WorkLocationHistoryActionGlobalPolicyRemoved WorkLocationHistoryAction = "global_policy_removed"
)

// IsValid checks if the work location history action is valid
func (a WorkLocationHistoryAction) IsValid() bool {
	switch a {
	case WorkLocationHistoryActionSelfSet,
		WorkLocationHistoryActionLeadOverride,
		WorkLocationHistoryActionAdminOverride,
		WorkLocationHistoryActionGlobalPolicyCreated,
		WorkLocationHistoryActionGlobalPolicyRemoved:
		return true
	}
	return false
}

// String returns the string representation of the work location history action
func (a WorkLocationHistoryAction) String() string {
	return string(a)
}

// DayStatus represents the status of a day
type DayStatus string

const (
	DayStatusNormal       DayStatus = "normal"
	DayStatusOfficeClosed DayStatus = "office_closed"
	DayStatusGovtHoliday  DayStatus = "govt_holiday"
	DayStatusCelebration  DayStatus = "celebration"
	DayStatusWeekend      DayStatus = "weekend"
)

// IsValid checks if the day status is valid
func (d DayStatus) IsValid() bool {
	switch d {
	case DayStatusNormal, DayStatusOfficeClosed, DayStatusGovtHoliday, DayStatusCelebration, DayStatusWeekend:
		return true
	}
	return false
}

// String returns the string representation of the day status
func (d DayStatus) String() string {
	return string(d)
}

// HistoryAction represents an action in the participation history
type HistoryAction string

const (
	HistoryActionOptedIn     HistoryAction = "opted_in"
	HistoryActionOptedOut    HistoryAction = "opted_out"
	HistoryActionOverrideIn  HistoryAction = "override_in"
	HistoryActionOverrideOut HistoryAction = "override_out"
)

// IsValid checks if the history action is valid
func (h HistoryAction) IsValid() bool {
	switch h {
	case HistoryActionOptedIn, HistoryActionOptedOut, HistoryActionOverrideIn, HistoryActionOverrideOut:
		return true
	}
	return false
}

// String returns the string representation of the history action
func (h HistoryAction) String() string {
	return string(h)
}
