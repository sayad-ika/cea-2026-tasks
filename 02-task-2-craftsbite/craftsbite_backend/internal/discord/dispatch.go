package discord

import "github.com/sayad-ika/craftsbite/internal/config"

// commandACL maps each command to the set of roles allowed to execute it.
var commandACL = map[string]map[string]struct{}{
	"meal":         {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"location":     {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"status":       {"employee": {}, "team_lead": {}, "admin": {}, "logistics": {}},
	"override":     {"team_lead": {}, "admin": {}},
	"team-summary": {"team_lead": {}, "admin": {}, "logistics": {}},
	"headcount":    {"admin": {}, "logistics": {}},
	"set-day":      {"admin": {}},
	"admin":        {"admin": {}},
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

func Dispatch(cfg *config.Config, commandName string) (string, bool) {
	switch commandName {
	case "meal", "location", "status":
		return cfg.LambdaSelfFunctionName, true
	case "override", "team-summary":
		return cfg.LambdaManagementFunctionName, true
	case "headcount", "set-day", "admin":
		return cfg.LambdaOpsFunctionName, true
	default:
		return "", false
	}
}
