package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CalendarUsecase interface {
	CreateCalendar(ctx context.Context, userID uuid.UUID, name, color, description string, projectID *uuid.UUID) (*repository.Calendar, error)
	ListCalendars(ctx context.Context, userID uuid.UUID) ([]repository.Calendar, error)
	DeleteCalendar(ctx context.Context, userID, calendarID uuid.UUID) error

	CreateEvent(ctx context.Context, userID, projectID uuid.UUID, title, description, location string, startAt, endAt time.Time, isAllDay bool) (*repository.ScheduledEvent, error)
	ListEvents(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]repository.ListEventsByRangeRow, error)

	CreateTimetableSlot(ctx context.Context, userID, projectID uuid.UUID, dayOfWeek int32, start, end time.Time, location string) (*repository.TimetableSlot, error)
	ListTimetable(ctx context.Context, userID uuid.UUID) ([]repository.ListTimetableSlotsRow, error)

	SyncDailySchedule(ctx context.Context, userID uuid.UUID, date time.Time) error
}

type calendarUsecase struct {
	repo      *repository.Queries
	txManager db.TxManager
}

func NewCalendarUsecase(repo *repository.Queries, txManager db.TxManager) CalendarUsecase {
	return &calendarUsecase{
		repo:      repo,
		txManager: txManager,
	}
}

func (u *calendarUsecase) CreateEvent(ctx context.Context, userID, projectID uuid.UUID, title, description, location string, startAt, endAt time.Time, isAllDay bool) (*repository.ScheduledEvent, error) {
	// デフォルトカレンダー
	defaultCal, err := u.repo.GetDefaultCalendar(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewNotFoundError("no calendar found for user")
		}
		return nil, fmt.Errorf("failed to get default calendar: %w", err)
	}
	defaultCalendarID := defaultCal.ID
	arg := repository.CreateEventParams{
		UserID:      userID,
		ProjectID:   projectID,
		CalendarID:  pgtype.UUID{Bytes: defaultCalendarID, Valid: true},
		Title:       title,
		Description: toTextFromStr(description),
		Location:    toTextFromStr(location),
		StartAt:     toTimestamp(&startAt),
		EndAt:       toTimestamp(&endAt),
		IsAllDay:    isAllDay,
		IcalUid:     pgtype.Text{String: uuid.NewString(), Valid: true},
		Status:      toTextFromStr("CONFIRMED"),
		Rrule:       pgtype.Text{Valid: false},
		Etag:        pgtype.Text{Valid: false},
		Sequence:    0,
	}
	event, err := u.repo.CreateEvent(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return &event, nil
}

func (u *calendarUsecase) ListEvents(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]repository.ListEventsByRangeRow, error) {
	events, err := u.repo.ListEventsByRange(ctx, repository.ListEventsByRangeParams{
		UserID:    userID,
		StartTime: toTimestamp(&start),
		EndTime:   toTimestamp(&end),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return events, nil
}

func (u *calendarUsecase) CreateTimetableSlot(ctx context.Context, userID, projectID uuid.UUID, dayOfWeek int32, start, end time.Time, location string) (*repository.TimetableSlot, error) {

	slot, err := u.repo.CreateTimetableSlot(ctx, repository.CreateTimetableSlotParams{
		UserID:    userID,
		ProjectID: projectID,
		DayOfWeek: int16(dayOfWeek),
		StartTime: toPgTime(start),
		EndTime:   toPgTime(end),
		Location:  toTextFromStr(location),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create slot: %w", err)
	}
	return &slot, nil
}

func (u *calendarUsecase) ListTimetable(ctx context.Context, userID uuid.UUID) ([]repository.ListTimetableSlotsRow, error) {
	slots, err := u.repo.ListTimetableSlots(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list timetable: %w", err)
	}
	return slots, nil
}

func (u *calendarUsecase) SyncDailySchedule(ctx context.Context, userID uuid.UUID, targetDate time.Time) error {
	dow := int16(targetDate.Weekday())

	return u.txManager.ReadCommitted(ctx, func(q *repository.Queries) error {
		slots, err := q.ListTimetableSlotsByDayOfWeek(ctx, repository.ListTimetableSlotsByDayOfWeekParams{
			UserID:    userID,
			DayOfWeek: dow,
		})
		if err != nil {
			return err
		}

		for _, slot := range slots {
			start := mergeDateAndTime(targetDate, slot.StartTime.Microseconds)
			end := mergeDateAndTime(targetDate, slot.EndTime.Microseconds)

			_, err := q.CreateTimeEntry(ctx, repository.CreateTimeEntryParams{
				UserID:    userID,
				ProjectID: slot.ProjectID,
				StartedAt: toTimestamp(&start),
				Note:      toTextFromStr("Auto-generated from Timetable"),
				EndedAt:   toTimestamp(&end),
			})
			if err != nil {
				return fmt.Errorf("failed to sync slot for project %s: %w", slot.ProjectID, err)
			}
		}
		return nil
	})
}

func toPgTime(t time.Time) pgtype.Time {
	base := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return pgtype.Time{Microseconds: t.Sub(base).Microseconds(), Valid: true}
}

func mergeDateAndTime(date time.Time, micros int64) time.Time {
	base := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return base.Add(time.Duration(micros) * time.Microsecond)
}
func (u *calendarUsecase) CreateCalendar(ctx context.Context, userID uuid.UUID, name, color, description string, projectID *uuid.UUID) (*repository.Calendar, error) {
	arg := repository.CreateCalendarParams{
		UserID:      userID,
		Name:        name,
		Color:       toTextFromStr(color),
		Description: toTextFromStr(description),
		ProjectID:   toUUID(projectID),
	}
	calendar, err := u.repo.CreateCalendar(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar: %w", err)
	}
	return &calendar, nil
}

func (u *calendarUsecase) ListCalendars(ctx context.Context, userID uuid.UUID) ([]repository.Calendar, error) {
	calendars, err := u.repo.ListCalendars(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}
	return calendars, nil
}

func (u *calendarUsecase) DeleteCalendar(ctx context.Context, userID, calendarID uuid.UUID) error {
	err := u.repo.DeleteCalendar(ctx, repository.DeleteCalendarParams{
		ID:     calendarID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete calendar: %w", err)
	}
	return nil
}
