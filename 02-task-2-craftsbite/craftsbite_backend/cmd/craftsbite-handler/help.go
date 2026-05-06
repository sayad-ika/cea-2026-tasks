package main

import (
	"context"
	"strings"

	appconfig "github.com/sayad-ika/craftsbite/internal/config"
	"github.com/sayad-ika/craftsbite/internal/discord"
	"github.com/sayad-ika/craftsbite/internal/gchat"
	"github.com/sayad-ika/craftsbite/internal/payload"
)

const (
	helpTitle    = "CraftsBite Help"
	helpSubtitle = "Only commands available to your role are shown."
)

type helpCommand struct {
	Usage       string
	Description string
}

type helpSection struct {
	Title    string
	Commands []helpCommand
}

func handleHelpCommand(ctx context.Context, cfg *appconfig.Config, event payload.CommandEvent) error {
	sections := helpSectionsForRole(event.Role)
	if event.Source == "gchat" {
		card, err := gchat.HelpCard(helpTitle, helpSubtitle, toGChatHelpSections(sections))
		if err != nil {
			return err
		}
		return sendGChatCard(ctx, cfg, event, card)
	}
	return sendDiscordMessage(ctx, cfg, event, buildDiscordHelpMessage(sections))
}

func helpSectionsForRole(role string) []helpSection {
	sections := []helpSection{personalHelpSection()}
	switch role {
	case "team_lead":
		sections = append(sections, teamLeadHelpSection())
	case "logistics":
		sections = append(sections, operationsHelpSection())
	case "admin":
		sections = append(sections, teamLeadHelpSection(), operationsHelpSection(), adminHelpSection())
	}
	return sections
}

func personalHelpSection() helpSection {
	return helpSection{
		Title: "Personal",
		Commands: []helpCommand{
			{Usage: "/help", Description: "Show the commands available to you."},
			{Usage: "/status [date]", Description: "View your current location and meal status."},
			{Usage: "/meal [in|out] [meal] [date]", Description: "Set a meal choice for a date."},
			{Usage: "/meal [meal] [date]", Description: "Toggle the current meal choice when status is omitted."},
			{Usage: "/location [office|wfh] [date]", Description: "Set your work location for a date."},
			{Usage: "/location [date]", Description: "Toggle between Office and WFH when location is omitted."},
		},
	}
}

func teamLeadHelpSection() helpSection {
	return helpSection{
		Title: "Team Lead",
		Commands: []helpCommand{
			{Usage: "/team-summary [date] [team_id]", Description: "View your team's meal and location summary. Admins can also pass team_id."},
			{Usage: "/override <target> <entry> <date> [meal] [value] <reason>", Description: "Override a team member's meal or location entry. In Google Chat, use `--` before the reason when you also pass meal or value."},
		},
	}
}

func operationsHelpSection() helpSection {
	return helpSection{
		Title: "Operations",
		Commands: []helpCommand{
			{Usage: "/headcount [date]", Description: "View meal-wise headcount totals and the office/WFH split."},
		},
	}
}

func adminHelpSection() helpSection {
	return helpSection{
		Title: "Admin",
		Commands: []helpCommand{
			{Usage: "/schedule-day <date> <status> [meals] [reason]", Description: "Configure the day schedule, meals, and optional note."},
		},
	}
}

func buildDiscordHelpMessage(sections []helpSection) discord.Message {
	fields := make([]discord.EmbedField, 0, len(sections))
	for _, section := range sections {
		fields = append(fields, discord.EmbedField{
			Name:  section.Title,
			Value: discordHelpSectionText(section),
		})
	}
	return discord.EmbedMessage(discord.BrandEmbed(helpTitle, helpSubtitle, fields))
}

func discordHelpSectionText(section helpSection) string {
	lines := make([]string, 0, len(section.Commands))
	for _, command := range section.Commands {
		lines = append(lines, "`"+command.Usage+"` - "+command.Description)
	}
	return strings.Join(lines, "\n")
}

func toGChatHelpSections(sections []helpSection) []gchat.HelpSection {
	out := make([]gchat.HelpSection, 0, len(sections))
	for _, section := range sections {
		commands := make([]gchat.HelpCommand, 0, len(section.Commands))
		for _, command := range section.Commands {
			commands = append(commands, gchat.HelpCommand{Usage: command.Usage, Description: command.Description})
		}
		out = append(out, gchat.HelpSection{Title: section.Title, Commands: commands})
	}
	return out
}
