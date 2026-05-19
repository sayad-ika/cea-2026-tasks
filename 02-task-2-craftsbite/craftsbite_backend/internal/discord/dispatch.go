package discord

// commandACL maps each command to the set of roles allowed to execute it.
var commandACL = map[string]map[string]struct{}{
	"help":         {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"init":         {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"meal":         {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"location":     {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"status":       {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"override":     {"team_lead": {}, "admin": {}},
	"team-summary": {"team_lead": {}, "admin": {}},
	"headcount":    {"admin": {}, "logistics": {}},
	"schedule-day": {"admin": {}},
	"admin":        {"admin": {}},
	"admin-init":   {"admin": {}},
}

// CheckPermission returns true if the given role is allowed to execute the command.
// Unknown commands are denied.
func CheckPermission(commandName, role string) bool {
	allowed, ok := commandACL[commandName]
	if !ok {
		return false
	}
	_, permitted := allowed[role]
	return permitted
}

func IsKnownCommand(commandName string) bool {
	_, ok := commandACL[commandName]
	return ok
}
