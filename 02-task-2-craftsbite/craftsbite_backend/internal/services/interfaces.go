package services

import (
	"context"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

type DayScheduleReader interface {
	GetDay(ctx context.Context, date string) (*repository.DaySchedule, error)
	GetAvailableMeals(ctx context.Context, date string) ([]string, error)
}

type DayScheduleWriter interface {
	DayScheduleReader
	UpsertDaySchedule(ctx context.Context, schedule repository.DaySchedule) error
	DeleteDaySchedule(ctx context.Context, date string) error
}

type ParticipationReader interface {
	GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]repository.MealParticipation, error)
	GetParticipationsByDate(ctx context.Context, date string) ([]repository.MealParticipation, error)
}

type ParticipationWriter interface {
	ParticipationReader
	UpsertParticipation(ctx context.Context, p repository.MealParticipation, prevUpdatedAt time.Time) error
}

type LocationReader interface {
	GetWorkLocation(ctx context.Context, userID, date string) (*repository.WorkLocation, error)
	GetWorkLocationsByDate(ctx context.Context, date string) ([]repository.WorkLocation, error)
}

type LocationWriter interface {
	LocationReader
	UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation, prevUpdatedAt time.Time) error
}

type UserReader interface {
	GetUserByDiscordID(ctx context.Context, discordID string) (userID, role string, err error)
	GetUserByGChatEmail(ctx context.Context, email string) (userID, role string, err error)
	ListActiveUsers(ctx context.Context) ([]repository.User, error)
	ListActiveUsersByRoles(ctx context.Context, roles ...string) ([]repository.User, error)
	GetUserByID(ctx context.Context, userID string) (*repository.User, error)
}

type TeamReader interface {
	GetTeamByID(ctx context.Context, teamID string) (*repository.Team, error)
	GetTeamMembers(ctx context.Context, teamID string) ([]repository.TeamMember, error)
	FindTeamsByLeadID(ctx context.Context, leadUserID string) ([]repository.Team, error)
}

type HeadcountStore interface {
	DayScheduleReader
	ParticipationReader
	LocationReader
	UserReader
	TeamReader
}

type TeamSummaryStore interface {
	DayScheduleReader
	ParticipationReader
	LocationReader
	UserReader
	TeamReader
}
