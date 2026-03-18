package payload

type CommandEvent struct {
	UserID           string                 `json:"userID"`
	Role             string                 `json:"role"`
	DiscordID        string                 `json:"discordId"`
	CommandName      string                 `json:"commandName"`
	Options          map[string]interface{} `json:"options"`
	InteractionToken string                 `json:"interactionToken"`
	ApplicationID    string                 `json:"applicationId"`
	Source           string                 `json:"source"`
	GChatSpaceName   string                 `json:"gchatSpaceName,omitempty"`
	GChatMessageName string                 `json:"gchatMessageName,omitempty"`
}
