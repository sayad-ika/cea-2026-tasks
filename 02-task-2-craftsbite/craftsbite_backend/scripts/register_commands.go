//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	optTypeString  = 3
	optTypeInteger = 4
	optTypeBoolean = 5
)

type commandOption struct {
	Type        int             `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Required    bool            `json:"required,omitempty"`
	Choices     []commandChoice `json:"choices,omitempty"`
}

type commandChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type slashCommand struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Options     []commandOption `json:"options,omitempty"`
}

func commands() []slashCommand {
	return []slashCommand{
		{
			Name:        "help",
			Description: "Show the commands available to you",
		},
		{
			Name:        "init",
			Description: "Open the interactive meal and location setup panel",
		},
		{
			Name:        "override",
			Description: "Override a team member's meal or location entry",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "target",
					Description: "Team member work email",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "entry",
					Description: "Which entry to override",
					Required:    true,
					Choices: []commandChoice{
						{Name: "Meal", Value: "meal"},
						{Name: "Location", Value: "location"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: today, tomorrow, +N, or YYYY-MM-DD",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "reason",
					Description: "Reason for the override",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "meal",
					Description: "Meal type (only for meal overrides; defaults to all)",
					Required:    false,
					Choices: []commandChoice{
						{Name: "Lunch", Value: "lunch"},
						{Name: "Snacks", Value: "snacks"},
						{Name: "Iftar", Value: "iftar"},
						{Name: "Event Dinner", Value: "event_dinner"},
						{Name: "Optional Dinner", Value: "optional_dinner"},
						{Name: "All", Value: "all"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "value",
					Description: "For meal: in|out. For location: office|wfh. Omit to toggle.",
					Required:    false,
				},
			},
		},
		{
			Name:        "override-init",
			Description: "Open the interactive override panel",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
			},
		},
		{
			Name:        "meal",
			Description: "Update your meal participation for a date, or toggle it if status is omitted",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "status",
					Description: "Participation status (omit to toggle current choice)",
					Required:    false,
					Choices: []commandChoice{
						{Name: "In", Value: "in"},
						{Name: "Out", Value: "out"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "meal",
					Description: "Meal type (defaults to all if omitted)",
					Required:    false,
					Choices: []commandChoice{
						{Name: "Lunch", Value: "lunch"},
						{Name: "Snacks", Value: "snacks"},
						{Name: "Iftar", Value: "iftar"},
						{Name: "Event Dinner", Value: "event_dinner"},
						{Name: "Optional Dinner", Value: "optional_dinner"},
						{Name: "All", Value: "all"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
			},
		},
		{
			Name:        "status",
			Description: "View your participation status for a date",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
			},
		},
		{
			Name:        "location",
			Description: "Set your work location for a date, or toggle it if location is omitted",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "location",
					Description: "Work location (omit to toggle current choice)",
					Required:    false,
					Choices: []commandChoice{
						{Name: "Office", Value: "office"},
						{Name: "WFH", Value: "wfh"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD, or week",
					Required:    false,
				},
			},
		},
		{
			Name:        "team-summary",
			Description: "View team participation summary for a date (Team Leads only)",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
				{
					Type:        optTypeString,
					Name:        "team_id",
					Description: "Team ID to query (Admin only; defaults to your own team)",
					Required:    false,
				},
			},
		},
		{
			Name:        "headcount",
			Description: "View org-wide headcount for a date (Admin/Logistics only)",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
			},
		},
		{
			Name:        "schedule-day",
			Description: "Configure a day's schedule and available meals (Admin only)",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date: YYYY-MM-DD, YYYY-MM-DD..YYYY-MM-DD",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "status",
					Description: "Day status",
					Required:    true,
					Choices: []commandChoice{
						{Name: "Normal Day", Value: "normal"},
						{Name: "Office Closed", Value: "office_closed"},
						{Name: "Government Holiday", Value: "govt_holiday"},
						{Name: "Celebration", Value: "celebration"},
						{Name: "Weekend", Value: "weekend"},
						{Name: "Event Day", Value: "event_day"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "meals",
					Description: "Comma-separated meal types (lunch,snacks,iftar,event_dinner,optional_dinner)",
					Required:    false,
				},
				{
					Type:        optTypeString,
					Name:        "reason",
					Description: "Optional reason or note",
					Required:    false,
				},
			},
		},
		{
			Name:        "admin-init",
			Description: "Open the interactive admin schedule setup panel",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Start date: tomorrow (default), today, +N, or YYYY-MM-DD",
					Required:    false,
				},
			},
		},
	}
}

func main() {
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: DISCORD_BOT_TOKEN is not set")
		os.Exit(1)
	}

	appID := os.Getenv("DISCORD_APPLICATION_ID")
	if appID == "" {
		fmt.Fprintln(os.Stderr, "error: DISCORD_APPLICATION_ID is not set")
		os.Exit(1)
	}

	cmds := commands()

	body, err := json.Marshal(cmds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to marshal commands: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("https://discord.com/api/v10/applications/%s/commands", appID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to build request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+token)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: HTTP request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read response body: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "error: Discord API returned %d\n%s\n", resp.StatusCode, respBody)
		os.Exit(1)
	}

	var registered []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(respBody, &registered); err != nil {
		fmt.Printf("registered %d command(s) (response decode failed: %v)\n", len(cmds), err)
		return
	}

	fmt.Printf("registered %d global slash command(s):\n", len(registered))
	for _, c := range registered {
		fmt.Printf("  %-16s  id=%s\n", "/"+c.Name, c.ID)
	}
}
