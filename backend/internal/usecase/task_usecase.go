package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TaskUsecase interface {
	CreateTask(ctx context.Context, userID, projectID uuid.UUID, title, note string, dueDate *time.Time) (*repository.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID, status *repository.TaskStatus, from, to *time.Time) ([]repository.ListTasksWithStatsRow, error)
	UpdateTaskStatus(ctx context.Context, userID, taskID uuid.UUID, status repository.TaskStatus) (*repository.Task, error)

	AddChecklistItem(ctx context.Context, taskID uuid.UUID, content string) (*repository.ChecklistItem, error)
	ToggleChecklistItem(ctx context.Context, itemID uuid.UUID, isCompleted bool) (*repository.ChecklistItem, error)
}

type taskUsecase struct {
	repo *repository.Queries
}

func NewTaskUsecase(repo *repository.Queries) TaskUsecase {
	return &taskUsecase{repo: repo}
}

func (u *taskUsecase) CreateTask(ctx context.Context, userID, projectID uuid.UUID, title, note string, dueDate *time.Time) (*repository.Task, error) {
	// Find default calendar for user
	var calendarID pgtype.UUID
	calendars, err := u.repo.ListCalendars(ctx, userID)
	if err == nil && len(calendars) > 0 {
		calendarID = pgtype.UUID{Bytes: calendars[0].ID, Valid: true}
	}

	arg := repository.CreateTaskParams{
		UserID:       userID,
		ProjectID:    projectID,
		Title:        title,
		NoteMarkdown: toTextFromStr(note),
		DueDate:      toTimestamp(dueDate),
		Priority:     pgtype.Int2{Int16: 0, Valid: true}, // Default: None
		Status:       repository.TaskStatusTODO,
		IcalUid:      pgtype.Text{String: uuid.NewString(), Valid: true},
		CalendarID:   calendarID,
	}

	task, err := u.repo.CreateTask(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return &task, nil
}

func (u *taskUsecase) ListTasks(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID, status *repository.TaskStatus, from, to *time.Time) ([]repository.ListTasksWithStatsRow, error) {
	var argStatus repository.NullTaskStatus
	if status != nil {
		argStatus = repository.NullTaskStatus{TaskStatus: *status, Valid: true}
	}

	arg := repository.ListTasksWithStatsParams{
		UserID:    userID,
		ProjectID: toUUID(projectID),
		Status:    argStatus,
		FromDate:  toTimestamp(from),
		ToDate:    toTimestamp(to),
	}

	tasks, err := u.repo.ListTasksWithStats(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

func (u *taskUsecase) UpdateTaskStatus(ctx context.Context, userID, taskID uuid.UUID, status repository.TaskStatus) (*repository.Task, error) {
	var completedAt pgtype.Timestamptz
	if status == repository.TaskStatusDONE {
		completedAt = toTimestamp(ptr(time.Now()))
	} else {
		completedAt = pgtype.Timestamptz{Valid: false}
	}
	arg := repository.UpdateTaskParams{
		ID:          taskID,
		UserID:      userID,
		Status:      status,
		CompletedAt: completedAt,
	}

	task, err := u.repo.UpdateTask(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return &task, nil
}

func (u *taskUsecase) AddChecklistItem(ctx context.Context, taskID uuid.UUID, content string) (*repository.ChecklistItem, error) {
	item, err := u.repo.CreateChecklistItem(ctx, repository.CreateChecklistItemParams{
		TaskID:  taskID,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add item: %w", err)
	}
	return &item, nil
}

func (u *taskUsecase) ToggleChecklistItem(ctx context.Context, itemID uuid.UUID, isCompleted bool) (*repository.ChecklistItem, error) {
	item, err := u.repo.UpdateChecklistItem(ctx, repository.UpdateChecklistItemParams{
		ID:          itemID,
		IsCompleted: pgtype.Bool{Bool: isCompleted, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to toggle item: %w", err)
	}
	return &item, nil
}
