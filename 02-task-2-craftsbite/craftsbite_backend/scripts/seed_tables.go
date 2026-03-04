//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"

	"craftsbite-backend/internal/dynamo"
)

const (
	// Teams
	teamSagaID  = "aaaaaaaa-0000-0000-0000-000000000001"
	teamMimirID = "bbbbbbbb-0000-0000-0000-000000000002"

	// Team leads
	userRafiqID  = "cccccccc-0000-0000-0000-000000000001" // team_lead — SAGA
	userNusratID = "dddddddd-0000-0000-0000-000000000002" // team_lead — MIMIR

	// SAGA members
	userArifID    = "eeeeeeee-0000-0000-0000-000000000003" // employee
	userSumaiyaID = "ffffffff-0000-0000-0000-000000000004" // employee

	// MIMIR members
	userTanvirID = "00000000-aaaa-0000-0000-000000000005" // employee
	userFatemaID = "00000000-bbbb-0000-0000-000000000006" // employee

	// Org-level roles
	userMahbubID  = "00000000-cccc-0000-0000-000000000007" // admin
	userSharminID = "00000000-dddd-0000-0000-000000000008" // logistics

	// Discord IDs
	discordRafiq   = "111111111111111111"
	discordNusrat  = "222222222222222222"
	discordArif    = "333333333333333333"
	discordSumaiya = "444444444444444444"
	discordTanvir  = "555555555555555555"
	discordFatema  = "666666666666666666"
	discordMahbub  = "777777777777777777"
	discordSharmin = "888888888888888888"

	usersTable = "craftsbite-users"
)

func main() {
	_ = godotenv.Load()

	client := dynamo.GetClient()
	ctx := context.Background()

	items := []map[string]types.AttributeValue{
		// --- Teams ---
		teamItem(teamSagaID, "SAGA", userRafiqID, true),
		teamListingItem(teamSagaID, true),
		teamItem(teamMimirID, "MIMIR", userNusratID, true),
		teamListingItem(teamMimirID, true),

		// --- Users ---

		// Team leads
		userProfileItem(userRafiqID, "rafiq@craftsbite.com", "Rafiq Hassan", "team_lead", true),
		userListingItem(userRafiqID, "team_lead", true),
		userProfileItem(userNusratID, "nusrat@craftsbite.com", "Nusrat Jahan", "team_lead", true),
		userListingItem(userNusratID, "team_lead", true),

		// SAGA employees
		userProfileItem(userArifID, "arif@craftsbite.com", "Arif Hossain", "employee", true),
		userListingItem(userArifID, "employee", true),
		userProfileItem(userSumaiyaID, "sumaiya@craftsbite.com", "Sumaiya Begum", "employee", true),
		userListingItem(userSumaiyaID, "employee", true),

		// MIMIR employees
		userProfileItem(userTanvirID, "tanvir@craftsbite.com", "Tanvir Ahmed", "employee", true),
		userListingItem(userTanvirID, "employee", true),
		userProfileItem(userFatemaID, "fatema@craftsbite.com", "Fatema Khatun", "employee", true),
		userListingItem(userFatemaID, "employee", true),

		// Admin
		userProfileItem(userMahbubID, "mahbub@craftsbite.com", "Mahbub Alam", "admin", true),
		userListingItem(userMahbubID, "admin", true),

		// Logistics
		userProfileItem(userSharminID, "sharmin@craftsbite.com", "Sharmin Akter", "logistics", true),
		userListingItem(userSharminID, "logistics", true),

		// --- TeamMember edges ---

		// SAGA: Rafiq (lead) + Arif + Sumaiya
		teamMemberItem(teamSagaID, userRafiqID),
		teamMemberItem(teamSagaID, userArifID),
		teamMemberItem(teamSagaID, userSumaiyaID),

		// MIMIR: Nusrat (lead) + Tanvir + Fatema
		teamMemberItem(teamMimirID, userNusratID),
		teamMemberItem(teamMimirID, userTanvirID),
		teamMemberItem(teamMimirID, userFatemaID),

		// --- Discord lookups ---
		discordLookupItem(discordRafiq, userRafiqID, "team_lead"),
		discordLookupItem(discordNusrat, userNusratID, "team_lead"),
		discordLookupItem(discordArif, userArifID, "employee"),
		discordLookupItem(discordSumaiya, userSumaiyaID, "employee"),
		discordLookupItem(discordTanvir, userTanvirID, "employee"),
		discordLookupItem(discordFatema, userFatemaID, "employee"),
		discordLookupItem(discordMahbub, userMahbubID, "admin"),
		discordLookupItem(discordSharmin, userSharminID, "logistics"),
	}

	for _, item := range items {
		if err := putItem(ctx, client, usersTable, item); err != nil {
			log.Fatalf("failed to put item %v: %v", item["PK"], err)
		}
	}

	fmt.Println("seed complete")
	printLookupTable()
}

// --- Item builders ---

func teamItem(id, name, leadID string, active bool) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#METADATA"},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"name":       &types.AttributeValueMemberS{Value: name},
		"teamLeadID": &types.AttributeValueMemberS{Value: leadID},
		"active":     &types.AttributeValueMemberBOOL{Value: active},
		// GSI1 — team lead lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "TEAMLEAD#" + leadID},
		"GSI1SK": &types.AttributeValueMemberS{Value: "TEAM#" + id},
	}
}

func teamListingItem(id string, active bool) map[string]types.AttributeValue {
	activeStr := boolPrefix(active)
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#LISTING"},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM_LISTING"},
		"id":         &types.AttributeValueMemberS{Value: id},
		// GSI1 — entity listing
		"GSI1PK": &types.AttributeValueMemberS{Value: "ENTITY#TEAM"},
		"GSI1SK": &types.AttributeValueMemberS{Value: activeStr + "#TEAM#" + id},
	}
}

func userProfileItem(id, email, name, role string, active bool) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "USER#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#PROFILE"},
		"entityType": &types.AttributeValueMemberS{Value: "USER"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"email":      &types.AttributeValueMemberS{Value: email},
		"name":       &types.AttributeValueMemberS{Value: name},
		"role":       &types.AttributeValueMemberS{Value: role},
		"active":     &types.AttributeValueMemberBOOL{Value: active},
		// GSI1 — email lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "EMAIL#" + email},
		"GSI1SK": &types.AttributeValueMemberS{Value: "USER#" + id},
	}
}

func userListingItem(id, role string, active bool) map[string]types.AttributeValue {
	activeStr := boolPrefix(active)
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "USER#" + id},
		"SK":         &types.AttributeValueMemberS{Value: "#LISTING"},
		"entityType": &types.AttributeValueMemberS{Value: "USER_LISTING"},
		"id":         &types.AttributeValueMemberS{Value: id},
		"role":       &types.AttributeValueMemberS{Value: role},
		// GSI1 — entity listing
		"GSI1PK": &types.AttributeValueMemberS{Value: "ENTITY#USER"},
		"GSI1SK": &types.AttributeValueMemberS{Value: activeStr + "#USER#" + id},
	}
}

func teamMemberItem(teamID, userID string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
		"SK":         &types.AttributeValueMemberS{Value: "MEMBER#" + userID},
		"entityType": &types.AttributeValueMemberS{Value: "TEAM_MEMBER"},
		"teamID":     &types.AttributeValueMemberS{Value: teamID},
		"userID":     &types.AttributeValueMemberS{Value: userID},
		// GSI1 — user → team reverse lookup
		"GSI1PK": &types.AttributeValueMemberS{Value: "USER_TEAMS#" + userID},
		"GSI1SK": &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
	}
}

func discordLookupItem(discordID, userID, role string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: "DISCORD#" + discordID},
		"SK":         &types.AttributeValueMemberS{Value: "#LOOKUP"},
		"entityType": &types.AttributeValueMemberS{Value: "DISCORD_LOOKUP"},
		"discordId":  &types.AttributeValueMemberS{Value: discordID},
		"userID":     &types.AttributeValueMemberS{Value: userID},
		"role":       &types.AttributeValueMemberS{Value: role},
	}
}

// --- Helpers ---

func putItem(ctx context.Context, client *dynamodb.Client, table string, item map[string]types.AttributeValue) error {
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	return err
}

func boolPrefix(active bool) string {
	if active {
		return "true"
	}
	return "false"
}

func printLookupTable() {
	fmt.Println("\n=== Discord ID → User Lookup ===")
	fmt.Printf("%-22s %-12s %-20s %-10s %s\n", "Discord ID", "Role", "Name", "Team", "User ID")
	fmt.Println(strings.Repeat("-", 88))

	rows := []struct {
		discord, role, name, team, userID string
	}{
		{discordRafiq,   "team_lead", "Rafiq Hassan",  "SAGA",  userRafiqID},
		{discordArif,    "employee",  "Arif Hossain",  "SAGA",  userArifID},
		{discordSumaiya, "employee",  "Sumaiya Begum", "SAGA",  userSumaiyaID},
		{discordNusrat,  "team_lead", "Nusrat Jahan",  "MIMIR", userNusratID},
		{discordTanvir,  "employee",  "Tanvir Ahmed",  "MIMIR", userTanvirID},
		{discordFatema,  "employee",  "Fatema Khatun", "MIMIR", userFatemaID},
		{discordMahbub,  "admin",     "Mahbub Alam",   "—",     userMahbubID},
		{discordSharmin, "logistics", "Sharmin Akter", "—",     userSharminID},
	}

	for _, r := range rows {
		fmt.Printf("%-22s %-12s %-20s %-10s %s\n", r.discord, r.role, r.name, r.team, r.userID)
	}
}