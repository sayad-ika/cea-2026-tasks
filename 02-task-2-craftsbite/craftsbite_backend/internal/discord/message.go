package discord

const (
	NoticeToneInfo    NoticeTone = "info"
	NoticeToneSuccess NoticeTone = "success"
	NoticeToneWarning NoticeTone = "warning"
	NoticeToneError   NoticeTone = "error"

	BrandColor = 0xF47621

	SuccessColor = 0x2F9E44
	WarningColor = 0xF08C00
	ErrorColor   = 0xE03131

	discordContentLimit = 2000
	maxDiscordEmbeds    = 10
)

type NoticeTone string

type Message struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
	Flags   int     `json:"flags,omitempty"`
}

type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type EmbedFooter struct {
	Text string `json:"text,omitempty"`
}

func NoticeMessage(title, description string) Message {
	return ToneMessage(title, description, NoticeToneInfo)
}

func ToneMessage(title, description string, tone NoticeTone) Message {
	return EmbedMessage(ToneEmbed(title, description, nil, tone))
}

func BrandEmbed(title, description string, fields []EmbedField) Embed {
	return Embed{
		Title:       title,
		Description: description,
		Color:       BrandColor,
		Fields:      fields,
		Footer:      &EmbedFooter{Text: "CraftsBite"},
	}
}

func ToneEmbed(title, description string, fields []EmbedField, tone NoticeTone) Embed {
	return Embed{
		Title:       title,
		Description: description,
		Color:       ColorForTone(tone),
		Fields:      fields,
		Footer:      &EmbedFooter{Text: "CraftsBite"},
	}
}

func EmbedMessage(embeds ...Embed) Message {
	return Message{Embeds: embeds}
}

func DefaultNoticeTitle(tone NoticeTone) string {
	switch tone {
	case NoticeToneSuccess:
		return "You're All Set"
	case NoticeToneWarning:
		return "Needs Your Attention"
	case NoticeToneError:
		return "We Couldn't Complete That"
	default:
		return "CraftsBite Update"
	}
}

func DefaultNoticeSubtitle(tone NoticeTone) string {
	switch tone {
	case NoticeToneSuccess:
		return "Your request was applied successfully"
	case NoticeToneWarning:
		return "Review the details below and try again"
	case NoticeToneError:
		return "Please try again in a moment"
	default:
		return "Latest status from your request"
	}
}

func ColorForTone(tone NoticeTone) int {
	switch tone {
	case NoticeToneSuccess:
		return SuccessColor
	case NoticeToneWarning:
		return WarningColor
	case NoticeToneError:
		return ErrorColor
	default:
		return BrandColor
	}
}

func NormalizeMessage(message Message) Message {
	if len(message.Content) > discordContentLimit {
		suffix := "\n_(message truncated)_"
		if len(suffix) >= discordContentLimit {
			suffix = "..."
		}
		message.Content = message.Content[:discordContentLimit-len(suffix)] + suffix
	}
	if len(message.Embeds) > maxDiscordEmbeds {
		message.Embeds = message.Embeds[:maxDiscordEmbeds]
	}
	return message
}
