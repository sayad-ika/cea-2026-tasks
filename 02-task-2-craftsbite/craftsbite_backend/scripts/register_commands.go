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
			Name:        "meal",
			Description: "Update your meal participation for a date",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date in YYYY-MM-DD format",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "status",
					Description: "Participation status",
					Required:    true,
					Choices: []commandChoice{
						{Name: "In", Value: "in"},
						{Name: "Out", Value: "out"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "meal",
					Description: "Meal type (defaults to both if omitted)",
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
			},
		},
		{
			Name:        "status",
			Description: "View your participation status for a date",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date in YYYY-MM-DD format",
					Required:    true,
				},
			},
		},
		{
			Name:        "location",
			Description: "Set your work location for a date",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date in YYYY-MM-DD format",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "location",
					Description: "Work location",
					Required:    true,
					Choices: []commandChoice{
						{Name: "Office", Value: "office"},
						{Name: "WFH", Value: "wfh"},
					},
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
					Description: "Date in YYYY-MM-DD format",
					Required:    true,
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
					Description: "Date in YYYY-MM-DD format",
					Required:    false,
				},
			},
		},
		{
			Name:        "set-day",
			Description: "Configure a day's schedule and available meals (Admin only)",
			Options: []commandOption{
				{
					Type:        optTypeString,
					Name:        "date",
					Description: "Date in YYYY-MM-DD format",
					Required:    true,
				},
				{
					Type:        optTypeString,
					Name:        "status",
					Description: "Day status",
					Required:    true,
					Choices: []commandChoice{
						{Name: "Normal", Value: "normal"},
						{Name: "Office Closed", Value: "office_closed"},
						{Name: "Government Holiday", Value: "govt_holiday"},
						{Name: "Celebration", Value: "celebration"},
						{Name: "Event Day", Value: "event_day"},
					},
				},
				{
					Type:        optTypeString,
					Name:        "meals",
					Description: "Comma-separated meal types available (e.g. lunch,snacks)",
					Required:    false,
				},
				{
					Type:        optTypeString,
					Name:        "reason",
					Description: "Optional note shown in bot replies for this day",
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

