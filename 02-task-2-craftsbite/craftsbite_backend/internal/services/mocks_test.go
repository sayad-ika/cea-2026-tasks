package services

import (
	"context"
	"time"

	"github.com/sayad-ika/craftsbite/internal/repository"
)

type mockDayScheduleReader struct {
	getDayFn            func(ctx context.Context, date string) (*repository.DaySchedule, error)
	getAvailableMealsFn func(ctx context.Context, date string) ([]string, error)
}

func (m *mockDayScheduleReader) GetDay(ctx context.Context, date string) (*repository.DaySchedule, error) {
	return m.getDayFn(ctx, date)
}

func (m *mockDayScheduleReader) GetAvailableMeals(ctx context.Context, date string) ([]string, error) {
	return m.getAvailableMealsFn(ctx, date)
}

type mockDayScheduleWriter struct {
	mockDayScheduleReader
	upsertFn func(ctx context.Context, schedule repository.DaySchedule) error
	deleteFn func(ctx context.Context, date string) error
}

func (m *mockDayScheduleWriter) UpsertDaySchedule(ctx context.Context, schedule repository.DaySchedule) error {
	return m.upsertFn(ctx, schedule)
}

func (m *mockDayScheduleWriter) DeleteDaySchedule(ctx context.Context, date string) error {
	return m.deleteFn(ctx, date)
}

type mockParticipationReader struct {
	getByUserDateFn func(ctx context.Context, userID, date string) ([]repository.MealParticipation, error)
	getByDateFn     func(ctx context.Context, date string) ([]repository.MealParticipation, error)
}

func (m *mockParticipationReader) GetParticipationsByUserDate(ctx context.Context, userID, date string) ([]repository.MealParticipation, error) {
	return m.getByUserDateFn(ctx, userID, date)
}

func (m *mockParticipationReader) GetParticipationsByDate(ctx context.Context, date string) ([]repository.MealParticipation, error) {
	return m.getByDateFn(ctx, date)
}

type mockParticipationWriter struct {
	mockParticipationReader
	upsertFn func(ctx context.Context, p repository.MealParticipation, prevUpdatedAt time.Time) error
}

func (m *mockParticipationWriter) UpsertParticipation(ctx context.Context, p repository.MealParticipation, prevUpdatedAt time.Time) error {
	return m.upsertFn(ctx, p, prevUpdatedAt)
}

type mockLocationReader struct {
	getFn         func(ctx context.Context, userID, date string) (*repository.WorkLocation, error)
	getByDateFn   func(ctx context.Context, date string) ([]repository.WorkLocation, error)
}

func (m *mockLocationReader) GetWorkLocation(ctx context.Context, userID, date string) (*repository.WorkLocation, error) {
	return m.getFn(ctx, userID, date)
}

func (m *mockLocationReader) GetWorkLocationsByDate(ctx context.Context, date string) ([]repository.WorkLocation, error) {
	return m.getByDateFn(ctx, date)
}

type mockLocationWriter struct {
	mockLocationReader
	upsertFn func(ctx context.Context, wl repository.WorkLocation, prevUpdatedAt time.Time) error
}

func (m *mockLocationWriter) UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation, prevUpdatedAt time.Time) error {
	return m.upsertFn(ctx, wl, prevUpdatedAt)
}

type mockUserReader struct {
	getByDiscordIDFn   func(ctx context.Context, discordID string) (string, string, error)
	getByGChatEmailFn  func(ctx context.Context, email string) (string, string, error)
	listActiveFn       func(ctx context.Context) ([]repository.User, error)
	listByRolesFn      func(ctx context.Context, roles ...string) ([]repository.User, error)
	getByIDFn          func(ctx context.Context, userID string) (*repository.User, error)
}

func (m *mockUserReader) GetUserByDiscordID(ctx context.Context, discordID string) (string, string, error) {
	return m.getByDiscordIDFn(ctx, discordID)
}

func (m *mockUserReader) GetUserByGChatEmail(ctx context.Context, email string) (string, string, error) {
	return m.getByGChatEmailFn(ctx, email)
}

func (m *mockUserReader) ListActiveUsers(ctx context.Context) ([]repository.User, error) {
	return m.listActiveFn(ctx)
}

func (m *mockUserReader) ListActiveUsersByRoles(ctx context.Context, roles ...string) ([]repository.User, error) {
	return m.listByRolesFn(ctx, roles...)
}

func (m *mockUserReader) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	return m.getByIDFn(ctx, userID)
}

type mockTeamReader struct {
	getByIDFn       func(ctx context.Context, teamID string) (*repository.Team, error)
	getMembersFn    func(ctx context.Context, teamID string) ([]repository.TeamMember, error)
	findByLeadIDFn  func(ctx context.Context, leadUserID string) ([]repository.Team, error)
}

func (m *mockTeamReader) GetTeamByID(ctx context.Context, teamID string) (*repository.Team, error) {
	return m.getByIDFn(ctx, teamID)
}

func (m *mockTeamReader) GetTeamMembers(ctx context.Context, teamID string) ([]repository.TeamMember, error) {
	return m.getMembersFn(ctx, teamID)
}

func (m *mockTeamReader) FindTeamsByLeadID(ctx context.Context, leadUserID string) ([]repository.Team, error) {
	return m.findByLeadIDFn(ctx, leadUserID)
}

type noErrStore struct {
	*mockDayScheduleReader
	*mockParticipationReader
	*mockLocationReader
	*mockUserReader
	*mockTeamReader
}

func (n *noErrStore) UpsertDaySchedule(ctx context.Context, schedule repository.DaySchedule) error                              { return nil }
func (n *noErrStore) DeleteDaySchedule(ctx context.Context, date string) error                                                  { return nil }
func (n *noErrStore) UpsertParticipation(ctx context.Context, p repository.MealParticipation, _ time.Time) error                 { return nil }
func (n *noErrStore) UpsertWorkLocation(ctx context.Context, wl repository.WorkLocation, _ time.Time) error                      { return nil }

func noErrDayReader(schedule *repository.DaySchedule, meals []string) *mockDayScheduleReader {
	return &mockDayScheduleReader{
		getDayFn:            func(_ context.Context, _ string) (*repository.DaySchedule, error) { return schedule, nil },
		getAvailableMealsFn: func(_ context.Context, _ string) ([]string, error) { return meals, nil },
	}
}

func noErrParticipationReader(records []repository.MealParticipation) *mockParticipationReader {
	return &mockParticipationReader{
		getByUserDateFn: func(_ context.Context, _, _ string) ([]repository.MealParticipation, error) { return records, nil },
		getByDateFn:     func(_ context.Context, _ string) ([]repository.MealParticipation, error) { return records, nil },
	}
}

func noErrLocationReader(wl *repository.WorkLocation) *mockLocationReader {
	return &mockLocationReader{
		getFn:       func(_ context.Context, _, _ string) (*repository.WorkLocation, error) { return wl, nil },
		getByDateFn: func(_ context.Context, _ string) ([]repository.WorkLocation, error) { return nil, nil },
	}
}

func noErrUserReader(users []repository.User) *mockUserReader {
	return &mockUserReader{
		getByDiscordIDFn:  func(_ context.Context, _ string) (string, string, error) { return "", "", nil },
		getByGChatEmailFn: func(_ context.Context, _ string) (string, string, error) { return "", "", nil },
		listActiveFn:      func(_ context.Context) ([]repository.User, error) { return users, nil },
		listByRolesFn:     func(_ context.Context, _ ...string) ([]repository.User, error) { return users, nil },
		getByIDFn: func(_ context.Context, _ string) (*repository.User, error) {
			if len(users) > 0 {
				return &users[0], nil
			}
			return nil, nil
		},
	}
}

func noErrTeamReader(teams []repository.Team, members []repository.TeamMember) *mockTeamReader {
	return &mockTeamReader{
		getByIDFn: func(_ context.Context, _ string) (*repository.Team, error) {
			if len(teams) > 0 {
				return &teams[0], nil
			}
			return nil, nil
		},
		getMembersFn:   func(_ context.Context, _ string) ([]repository.TeamMember, error) { return members, nil },
		findByLeadIDFn: func(_ context.Context, _ string) ([]repository.Team, error) { return teams, nil },
	}
}
