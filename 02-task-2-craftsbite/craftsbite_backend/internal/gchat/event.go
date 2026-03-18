package gchat

// Event is the top-level payload for Google Chat App Commands API events.
type Event struct {
	Chat ChatEvent `json:"chat"`
}

type ChatEvent struct {
	User              Sender             `json:"user"`
	EventTime         string             `json:"eventTime"`
	AppCommandPayload *AppCommandPayload `json:"appCommandPayload,omitempty"`
}

type AppCommandPayload struct {
	AppCommandMetadata AppCommandMetadata `json:"appCommandMetadata"`
	Space              Space              `json:"space"`
	Message            *Message           `json:"message,omitempty"`
}

type AppCommandMetadata struct {
	AppCommandID   float64 `json:"appCommandId"`
	AppCommandType string  `json:"appCommandType"`
}

type Message struct {
	Name         string        `json:"name"`
	Sender       Sender        `json:"sender"`
	Text         string        `json:"text"`
	SlashCommand *SlashCommand `json:"slashCommand,omitempty"`
}

type Sender struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Type        string `json:"type"`
}

type Space struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	SpaceType string `json:"spaceType"`
}

type SlashCommand struct {
	CommandID float64 `json:"commandId"`
}
