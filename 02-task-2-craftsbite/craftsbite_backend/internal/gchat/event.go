package gchat

// Event is the top-level payload for Google Chat App Commands API events.
type Event struct {
	Chat              ChatEvent         `json:"chat"`
	CommonEventObject CommonEventObject `json:"commonEventObject,omitempty"`
}

type ChatEvent struct {
	User                 Sender                `json:"user"`
	Space                Space                 `json:"space,omitempty"`
	EventTime            string                `json:"eventTime"`
	AppCommandPayload    *AppCommandPayload    `json:"appCommandPayload,omitempty"`
	ButtonClickedPayload *ButtonClickedPayload `json:"buttonClickedPayload,omitempty"`
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
	ArgumentText string        `json:"argumentText"`
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

type ButtonClickedPayload struct {
	IsDialogEvent   bool   `json:"isDialogEvent"`
	DialogEventType string `json:"dialogEventType"`
}

type CommonEventObject struct {
	InvokedFunction string               `json:"invokedFunction,omitempty"`
	Parameters      map[string]string    `json:"parameters,omitempty"`
	FormInputs      map[string]FormInput `json:"formInputs,omitempty"`
}

type FormInput struct {
	StringInputs *StringInputs `json:"stringInputs,omitempty"`
	DateInput    *DateInput    `json:"dateInput,omitempty"`
}

type StringInputs struct {
	Value []string `json:"value,omitempty"`
}

type DateInput struct {
	MsSinceEpoch int64 `json:"msSinceEpoch,omitempty"`
}
