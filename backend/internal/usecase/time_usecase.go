package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
)

type TimeUsecase interface {
	StartTimeEntry(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID, note string) (*repository.TimeEntry, error)
	StopTimeEntry(ctx context.Context, userID, entryID uuid.UUID) (*repository.TimeEntry, error)
	ListTimeEntries(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.ListTimeEntriesRow, error)
	GetContributionStats(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.GetGrowthStatsRow, error)
}

type timeUsecase struct {
	repo *repository.Queries
}

func NewTimeUsecase(repo *repository.Queries) TimeUsecase {
	return &timeUsecase{repo: repo}
}

func (u *timeUsecase) StartTimeEntry(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID, note string) (*repository.TimeEntry, error) {
	arg := repository.CreateTimeEntryParams{
		UserID:    userID,
		ProjectID: projectID,
		TaskID:    toUUID(taskID),
		StartedAt: toTimestamp(&time.Time{}),
		Note:      toTextFromStr(note),
	}
	arg.StartedAt = toTimestamp(ptr(time.Now()))

	entry, err := u.repo.CreateTimeEntry(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to start timer: %w", err)
	}
	return &entry, nil
}

func (u *timeUsecase) StopTimeEntry(ctx context.Context, userID, entryID uuid.UUID) (*repository.TimeEntry, error) {
	now := toTimestamp(ptr(time.Now()))
	entry, err := u.repo.StopTimeEntry(ctx, repository.StopTimeEntryParams{
		ID:      entryID,
		UserID:  userID,
		EndedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stop timer: %w", err)
	}
	return &entry, nil
}

func (u *timeUsecase) ListTimeEntries(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.ListTimeEntriesRow, error) {
	arg := repository.ListTimeEntriesParams{
		UserID:   userID,
		FromDate: toTimestamp(&from),
		ToDate:   toTimestamp(&to),
	}
	entries, err := u.repo.ListTimeEntries(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	return entries, nil
}

func (u *timeUsecase) GetContributionStats(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]repository.GetGrowthStatsRow, error) {
	stats, err := u.repo.GetGrowthStats(ctx, repository.GetGrowthStatsParams{
		UserID:   userID,
		FromDate: toTimestamp(&from),
		ToDate:   toTimestamp(&to),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

func ptr(t time.Time) *time.Time { return &t }
