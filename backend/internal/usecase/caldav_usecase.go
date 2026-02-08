package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CalDavUsecase interface {
	ExportCalendarToICal(ctx context.Context, userID, calendarID uuid.UUID) (string, error)
	ExportEventToICal(ctx context.Context, userID uuid.UUID, icalUID string) (string, error)
	ExportTaskToICal(ctx context.Context, userID uuid.UUID, icalUID string) (string, error)

	ImportFromICal(ctx context.Context, userID, calendarID uuid.UUID, icalData string) error
	DeleteResource(ctx context.Context, userID uuid.UUID, icalUID string) error

	GetCalendars(ctx context.Context, userID uuid.UUID) ([]repository.Calendar, error)
	GetCalendar(ctx context.Context, userID, calendarID uuid.UUID) (*repository.Calendar, error)

	GetEventsByRange(ctx context.Context, userID, calendarID uuid.UUID, start, end time.Time) ([]repository.ScheduledEvent, error)
	GetTasksByRange(ctx context.Context, userID, calendarID uuid.UUID, start, end time.Time) ([]repository.Task, error)
}

type calDavUsecase struct {
	repo      *repository.Queries
	txManager db.TxManager
}

func NewCalDavUsecase(repo *repository.Queries, txManager db.TxManager) CalDavUsecase {
	return &calDavUsecase{
		repo:      repo,
		txManager: txManager,
	}
}

func (u *calDavUsecase) DeleteResource(ctx context.Context, userID uuid.UUID, icalUID string) error {
	// Try event first
	err := u.repo.DeleteEventByICalUID(ctx, repository.DeleteEventByICalUIDParams{
		UserID:  userID,
		IcalUid: toTextFromStr(icalUID),
	})
	if err == nil {
		return nil
	}
	// Try task
	return u.repo.DeleteTaskByICalUID(ctx, repository.DeleteTaskByICalUIDParams{
		UserID:  userID,
		IcalUid: toTextFromStr(icalUID),
	})
}

func (u *calDavUsecase) GetCalendars(ctx context.Context, userID uuid.UUID) ([]repository.Calendar, error) {
	return u.repo.ListCalendars(ctx, userID)
}

func (u *calDavUsecase) GetCalendar(ctx context.Context, userID, calendarID uuid.UUID) (*repository.Calendar, error) {
	cal, err := u.repo.GetCalendar(ctx, repository.GetCalendarParams{
		ID:     calendarID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &cal, nil
}

func (u *calDavUsecase) GetEventsByRange(ctx context.Context, userID, calendarID uuid.UUID, start, end time.Time) ([]repository.ScheduledEvent, error) {
	if !start.IsZero() || !end.IsZero() {
		return u.repo.ListEventsByCalendarAndRange(ctx, repository.ListEventsByCalendarAndRangeParams{
			UserID:     userID,
			CalendarID: toUUID(&calendarID),
			StartTime:  toTimestamp(&start),
			EndTime:    toTimestamp(&end),
		})
	}
	return u.repo.ListEventsByCalendar(ctx, repository.ListEventsByCalendarParams{
		UserID:     userID,
		CalendarID: toUUID(&calendarID),
	})
}

func (u *calDavUsecase) GetTasksByRange(ctx context.Context, userID, calendarID uuid.UUID, start, end time.Time) ([]repository.Task, error) {
	if !start.IsZero() || !end.IsZero() {
		return u.repo.ListTasksByCalendarAndRange(ctx, repository.ListTasksByCalendarAndRangeParams{
			UserID:     userID,
			CalendarID: toUUID(&calendarID),
			StartTime:  toTimestamp(&start),
			EndTime:    toTimestamp(&end),
		})
	}
	return u.repo.ListTasksByCalendar(ctx, repository.ListTasksByCalendarParams{
		UserID:     userID,
		CalendarID: toUUID(&calendarID),
	})
}

func (u *calDavUsecase) ExportCalendarToICal(ctx context.Context, userID, calendarID uuid.UUID) (string, error) {
	events, err := u.repo.ListEventsByCalendar(ctx, repository.ListEventsByCalendarParams{
		UserID:     userID,
		CalendarID: toUUID(&calendarID),
	})
	if err != nil {
		return "", err
	}

	tasks, err := u.repo.ListTasksByCalendar(ctx, repository.ListTasksByCalendarParams{
		UserID:     userID,
		CalendarID: toUUID(&calendarID),
	})
	if err != nil {
		return "", err
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//Taskalyst//EN")
	cal.Props.SetText(ical.PropVersion, "2.0")

	for _, e := range events {
		event := eventToVEvent(&e)
		cal.Children = append(cal.Children, event.Component)
	}

	for _, t := range tasks {
		todo := taskToVTodo(&t)
		cal.Children = append(cal.Children, todo)
	}

	var sb strings.Builder
	if err := ical.NewEncoder(&sb).Encode(cal); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (u *calDavUsecase) ExportEventToICal(ctx context.Context, userID uuid.UUID, icalUID string) (string, error) {
	e, err := u.repo.GetEventByICalUID(ctx, repository.GetEventByICalUIDParams{
		UserID:  userID,
		IcalUid: toTextFromStr(icalUID),
	})
	if err != nil {
		return "", err
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//Taskalyst//EN")
	cal.Props.SetText(ical.PropVersion, "2.0")

	event := eventToVEvent(&e)
	cal.Children = append(cal.Children, event.Component)

	var sb strings.Builder
	if err := ical.NewEncoder(&sb).Encode(cal); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (u *calDavUsecase) ExportTaskToICal(ctx context.Context, userID uuid.UUID, icalUID string) (string, error) {
	t, err := u.repo.GetTaskByICalUID(ctx, repository.GetTaskByICalUIDParams{
		UserID:  userID,
		IcalUid: toTextFromStr(icalUID),
	})
	if err != nil {
		return "", err
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//Taskalyst//EN")
	cal.Props.SetText(ical.PropVersion, "2.0")

	todo := taskToVTodo(&t)
	cal.Children = append(cal.Children, todo)

	var sb strings.Builder
	if err := ical.NewEncoder(&sb).Encode(cal); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (u *calDavUsecase) ImportFromICal(ctx context.Context, userID, calendarID uuid.UUID, icalData string) error {
	dec := ical.NewDecoder(strings.NewReader(icalData))

	// Get calendar to check if it has a linked project
	calInfo, err := u.repo.GetCalendar(ctx, repository.GetCalendarParams{
		ID:     calendarID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to get calendar: %w", err)
	}

	var targetProjectID uuid.UUID
	if calInfo.ProjectID.Valid {
		targetProjectID = calInfo.ProjectID.Bytes
	} else {
		// Fallback to default project
		defaultProject, err := u.repo.GetDefaultProject(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return NewNotFoundError("no project found to import data")
			}
			return fmt.Errorf("failed to get default project: %w", err)
		}
		targetProjectID = defaultProject.ID
	}

	return u.txManager.ReadCommitted(ctx, func(q *repository.Queries) error {
		for {
			cal, err := dec.Decode()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			for _, event := range cal.Events() {
				uid, _ := event.Props.Text(ical.PropUID)
				summary, _ := event.Props.Text(ical.PropSummary)
				description, _ := event.Props.Text(ical.PropDescription)
				location, _ := event.Props.Text(ical.PropLocation)
				start, _ := event.Props.DateTime(ical.PropDateTimeStart, time.UTC)
				end, _ := event.Props.DateTime(ical.PropDateTimeEnd, time.UTC)

				// Check if exists
				_, err := q.GetEventByICalUID(ctx, repository.GetEventByICalUIDParams{
					UserID:  userID,
					IcalUid: toTextFromStr(uid),
				})

				if err == nil {
					// Update
					_, err = q.UpdateEventByICalUID(ctx, repository.UpdateEventByICalUIDParams{
						UserID:      userID,
						IcalUid:     toTextFromStr(uid),
						Title:       toTextFromStr(summary),
						Description: toTextFromStr(description),
						Location:    toTextFromStr(location),
						StartAt:     toTimestamp(&start),
						EndAt:       toTimestamp(&end),
					})
				} else {
					// Create
					_, err = q.CreateEvent(ctx, repository.CreateEventParams{
						UserID:      userID,
						ProjectID:   targetProjectID,
						CalendarID:  toUUID(&calendarID),
						Title:       summary,
						Description: toTextFromStr(description),
						Location:    toTextFromStr(location),
						StartAt:     toTimestamp(&start),
						EndAt:       toTimestamp(&end),
						IsAllDay:    false,
						IcalUid:     toTextFromStr(uid),
						Status:      toTextFromStr("CONFIRMED"),
					})
				}
				if err != nil {
					return err
				}
			}

			for _, child := range cal.Children {
				if child.Name != ical.CompToDo {
					continue
				}
				uid, _ := child.Props.Text(ical.PropUID)
				summary, _ := child.Props.Text(ical.PropSummary)
				description, _ := child.Props.Text(ical.PropDescription)
				due, _ := child.Props.DateTime(ical.PropDue, time.UTC)
				status, _ := child.Props.Text(ical.PropStatus)

				_, err := q.GetTaskByICalUID(ctx, repository.GetTaskByICalUIDParams{
					UserID:  userID,
					IcalUid: toTextFromStr(uid),
				})

				if err == nil {
					// Update
					_, err = q.UpdateTaskByICalUID(ctx, repository.UpdateTaskByICalUIDParams{
						UserID:       userID,
						IcalUid:      toTextFromStr(uid),
						Title:        toTextFromStr(summary),
						NoteMarkdown: toTextFromStr(description),
						DueDate:      toTimestamp(&due),
						Status:       repository.NullTaskStatus{TaskStatus: icalStatusToTaskStatus(status), Valid: true},
					})
				} else {
					// Create
					_, err = q.CreateTask(ctx, repository.CreateTaskParams{
						UserID:       userID,
						ProjectID:    targetProjectID,
						Title:        summary,
						NoteMarkdown: toTextFromStr(description),
						DueDate:      toTimestamp(&due),
						Priority:     pgtype.Int2{Int16: 0, Valid: true},
						CalendarID:   toUUID(&calendarID),
						IcalUid:      toTextFromStr(uid),
						Status:       icalStatusToTaskStatus(status),
					})
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func eventToVEvent(e *repository.ScheduledEvent) *ical.Event {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, e.IcalUid.String)
	event.Props.SetText(ical.PropSummary, e.Title)
	if e.Description.Valid {
		event.Props.SetText(ical.PropDescription, e.Description.String)
	}
	if e.Location.Valid {
		event.Props.SetText(ical.PropLocation, e.Location.String)
	}
	event.Props.SetDateTime(ical.PropDateTimeStart, e.StartAt.Time)
	event.Props.SetDateTime(ical.PropDateTimeEnd, e.EndAt.Time)
	if e.Status.Valid {
		event.Props.SetText(ical.PropStatus, e.Status.String)
	}
	return event
}

func taskToVTodo(t *repository.Task) *ical.Component {
	todo := ical.NewComponent(ical.CompToDo)
	todo.Props.SetText(ical.PropUID, t.IcalUid.String)
	todo.Props.SetText(ical.PropSummary, t.Title)
	if t.NoteMarkdown.Valid {
		todo.Props.SetText(ical.PropDescription, t.NoteMarkdown.String)
	}
	if t.DueDate.Valid {
		todo.Props.SetDateTime(ical.PropDue, t.DueDate.Time)
	}
	todo.Props.SetText(ical.PropStatus, taskStatusToICalStatus(t.Status))
	if t.CompletedAt.Valid {
		todo.Props.SetDateTime(ical.PropCompleted, t.CompletedAt.Time)
	}
	return todo
}

func taskStatusToICalStatus(s repository.TaskStatus) string {
	switch s {
	case repository.TaskStatusDONE:
		return "COMPLETED"
	case repository.TaskStatusDOING:
		return "IN-PROCESS"
	default:
		return "NEEDS-ACTION"
	}
}

func icalStatusToTaskStatus(s string) repository.TaskStatus {
	switch s {
	case "COMPLETED":
		return repository.TaskStatusDONE
	case "IN-PROCESS":
		return repository.TaskStatusDOING
	default:
		return repository.TaskStatusTODO
	}
}
