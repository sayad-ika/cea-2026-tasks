package repository

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Store struct {
	client *dynamodb.Client
	table  string
}

func NewStore(client *dynamodb.Client, table string) *Store {
	return &Store{client: client, table: table}
}

func (s *Store) GetDay(ctx context.Context, date string) (*DaySchedule, error) {
	return GetDay(ctx, s.client, s.table, date)
}

func (s *Store) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
	return GetAvailableMeals(ctx, s.client, s.table, date)
}

func (s *Store) UpsertDaySchedule(ctx context.Context, schedule DaySchedule) error {
	return UpsertDaySchedule(ctx, s.client, s.table, schedule)
}

func (s *Store) DeleteDaySchedule(ctx context.Context, date string) error {
	return DeleteDaySchedule(ctx, s.client, s.table, date)
}

func (s *Store) GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]MealParticipation, error) {
	return GetParticipationsByUserDate(ctx, s.client, s.table, userID, date)
}

func (s *Store) GetParticipationsByDate(ctx context.Context, date string) ([]MealParticipation, error) {
	return GetParticipationsByDate(ctx, s.client, s.table, date)
}

func (s *Store) UpsertParticipation(ctx context.Context, p MealParticipation, prevUpdatedAt time.Time) error {
	return UpsertParticipation(ctx, s.client, s.table, p, prevUpdatedAt)
}

func (s *Store) GetWorkLocation(ctx context.Context, userID, date string) (*WorkLocation, error) {
	return GetWorkLocation(ctx, s.client, s.table, userID, date)
}

func (s *Store) GetWorkLocationsByDate(ctx context.Context, date string) ([]WorkLocation, error) {
	return GetWorkLocationsByDate(ctx, s.client, s.table, date)
}

func (s *Store) UpsertWorkLocation(ctx context.Context, wl WorkLocation, prevUpdatedAt time.Time) error {
	return UpsertWorkLocation(ctx, s.client, s.table, wl, prevUpdatedAt)
}

func (s *Store) GetUserByDiscordID(ctx context.Context, discordID string) (string, string, error) {
	return GetUserByDiscordID(ctx, s.client, s.table, discordID)
}

func (s *Store) GetUserByGChatEmail(ctx context.Context, email string) (string, string, error) {
	return GetUserByGChatEmail(ctx, s.client, s.table, email)
}

func (s *Store) ListActiveUsers(ctx context.Context) ([]User, error) {
	return ListActiveUsers(ctx, s.client, s.table)
}

func (s *Store) ListActiveUsersByRoles(ctx context.Context, roles ...string) ([]User, error) {
	return ListActiveUsersByRoles(ctx, s.client, s.table, roles...)
}

func (s *Store) GetUserByID(ctx context.Context, userID string) (*User, error) {
	return GetUserByID(ctx, s.client, s.table, userID)
}

func (s *Store) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return GetTeamByID(ctx, s.client, s.table, teamID)
}

func (s *Store) GetTeamMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	return GetTeamMembers(ctx, s.client, s.table, teamID)
}

func (s *Store) FindTeamsByLeadID(ctx context.Context, leadUserID string) ([]Team, error) {
	return FindTeamsByLeadID(ctx, s.client, s.table, leadUserID)
}

func (s *Store) WriteAuditEntry(ctx context.Context, entry AuditEntry) error {
	return WriteAuditEntry(ctx, s.client, s.table, entry)
}
