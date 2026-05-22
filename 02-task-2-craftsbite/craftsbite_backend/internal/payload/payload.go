package payload

import "encoding/json"

type CommandEvent struct {
	UserID           string          `json:"userID"`
	Role             string          `json:"role"`
	DiscordID        string          `json:"discordId"`
	CommandName      string          `json:"commandName"`
	Options          json.RawMessage `json:"options"`
	InteractionToken string          `json:"interactionToken"`
	ApplicationID    string          `json:"applicationId"`
	Source           string          `json:"source"`
	GChatSpaceName   string          `json:"gchatSpaceName,omitempty"`
	GChatMessageName string          `json:"gchatMessageName,omitempty"`
	GChatViewerName  string          `json:"gchatViewerName,omitempty"`
}

func (e CommandEvent) ParseOptions(v interface{}) error {
	return json.Unmarshal(e.Options, v)
}

type MealOptions struct {
	Status string `json:"status"`
	Meal   string `json:"meal"`
	Date   string `json:"date"`
}

type InitOptions struct {
	Date     string   `json:"date"`
	Dates    []string `json:"dates"`
	Action   string   `json:"action"`
	Location string   `json:"location"`
	Meals    []string `json:"meals"`
}

type LocationOptions struct {
	Location string `json:"location"`
	Date     string `json:"date"`
}

type StatusOptions struct {
	Date string `json:"date"`
}

type HeadcountOptions struct {
	Date string `json:"date"`
}

type ScheduleDayOptions struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Meals  string `json:"meals"`
	Reason string `json:"reason"`
}

type AdminInitOptions struct {
	Action   string   `json:"action"`
	Date     string   `json:"date"`
	EndDate  string   `json:"end_date"`
	UseRange bool     `json:"use_range"`
	Status   string   `json:"status"`
	Meals    []string `json:"meals"`
	Reason   string   `json:"reason"`
}

type TeamSummaryOptions struct {
	Date   string `json:"date"`
	TeamID string `json:"team_id"`
}

type OverrideOptions struct {
	Target string `json:"target"`
	Entry  string `json:"entry"`
	Date   string `json:"date"`
	Meal   string `json:"meal"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type OverrideInitOptions struct {
	Action     string   `json:"action"`
	Date       string   `json:"date"`
	TeamID     string   `json:"team_id"`
	TargetMode string   `json:"target_mode"`
	Members    []string `json:"members"`
	Entry      string   `json:"entry"`
	Meal       string   `json:"meal"`
	Value      string   `json:"value"`
	Reason     string   `json:"reason"`
}
