package main

import (
	"strings"
	"testing"
)

func TestHelpSectionsForRole(t *testing.T) {
	tests := []struct {
		role    string
		want    []string
		notWant []string
	}{
		{role: "employee", want: []string{"Personal"}, notWant: []string{"Team Lead", "Operations", "Admin"}},
		{role: "team_lead", want: []string{"Personal", "Team Lead"}, notWant: []string{"Operations", "Admin"}},
		{role: "logistics", want: []string{"Personal", "Operations"}, notWant: []string{"Team Lead", "Admin"}},
		{role: "admin", want: []string{"Personal", "Team Lead", "Operations", "Admin"}},
	}

	for _, tt := range tests {
		sections := helpSectionsForRole(tt.role)
		for _, title := range tt.want {
			if !hasHelpSection(sections, title) {
				t.Fatalf("role %q missing section %q", tt.role, title)
			}
		}
		for _, title := range tt.notWant {
			if hasHelpSection(sections, title) {
				t.Fatalf("role %q unexpectedly included section %q", tt.role, title)
			}
		}
	}
}

func TestBuildDiscordHelpMessage(t *testing.T) {
	msg := buildDiscordHelpMessage(helpSectionsForRole("admin"))
	if len(msg.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(msg.Embeds))
	}
	embed := msg.Embeds[0]
	if embed.Title != helpTitle {
		t.Fatalf("title = %q, want %q", embed.Title, helpTitle)
	}
	if embed.Description != helpSubtitle {
		t.Fatalf("description = %q, want %q", embed.Description, helpSubtitle)
	}
	if len(embed.Fields) != 4 {
		t.Fatalf("expected 4 help sections, got %d", len(embed.Fields))
	}
	if embed.Fields[0].Name != "Personal" {
		t.Fatalf("first section = %q, want Personal", embed.Fields[0].Name)
	}
	if embed.Fields[3].Name != "Admin" {
		t.Fatalf("last section = %q, want Admin", embed.Fields[3].Name)
	}
	if !strings.Contains(embed.Fields[1].Value, "/override") {
		t.Fatal("expected team lead help section to include /override")
	}
	if strings.Contains(embed.Fields[3].Value, "/admin-init") {
		t.Fatal("discord admin help must not include gchat-only /admin-init")
	}
}

func TestGChatAdminHelpIncludesAdminInit(t *testing.T) {
	sections := helpSectionsForRoleForSource("admin", "gchat")
	admin := sections[len(sections)-1]
	if admin.Title != "Admin" {
		t.Fatalf("last section = %q, want Admin", admin.Title)
	}
	if !strings.Contains(discordHelpSectionText(admin), "/admin-init") {
		t.Fatal("expected gchat admin help section to include /admin-init")
	}
}

func hasHelpSection(sections []helpSection, title string) bool {
	for _, section := range sections {
		if section.Title == title {
			return true
		}
	}
	return false
}
