package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ResultUsecase interface {
	CreateResult(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID, resultType string, value float64, recordedAt time.Time, note string) (*repository.Result, error)
	ListResults(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID, from, to time.Time) ([]repository.ListResultsRow, error)
	DeleteResult(ctx context.Context, userID, resultID uuid.UUID) error
}

type resultUsecase struct {
	repo      *repository.Queries
	txManager db.TxManager
}

func NewResultUsecase(repo *repository.Queries, txManager db.TxManager) ResultUsecase {
	return &resultUsecase{
		repo:      repo,
		txManager: txManager,
	}
}
func toNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(fmt.Sprintf("%f", f))
	return n
}
func (u *resultUsecase) CreateResult(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID, resultType string, value float64, recordedAt time.Time, note string) (*repository.Result, error) {
	arg := repository.CreateResultParams{
		UserID:       userID,
		ProjectID:    projectID,
		TargetTaskID: toUUID(taskID),
		Type:         resultType,
		Value:        toNumeric(value),
		RecordedAt:   toTimestamp(&recordedAt),
		Note:         toTextFromStr(note),
	}

	result, err := u.repo.CreateResult(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create result: %w", err)
	}
	return &result, nil
}

func (u *resultUsecase) ListResults(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID, from, to time.Time) ([]repository.ListResultsRow, error) {
	arg := repository.ListResultsParams{
		UserID:    userID,
		ProjectID: toUUID(projectID),
		FromDate:  toTimestamp(&from),
		ToDate:    toTimestamp(&to),
	}

	results, err := u.repo.ListResults(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list results: %w", err)
	}
	return results, nil
}

func (u *resultUsecase) DeleteResult(ctx context.Context, userID, resultID uuid.UUID) error {
	err := u.repo.DeleteResult(ctx, repository.DeleteResultParams{
		ID:     resultID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete result: %w", err)
	}
	return nil
}
