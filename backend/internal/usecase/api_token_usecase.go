package usecase

import (
	"context"
	"fmt"

	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ApiTokenUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, name string) (string, *repository.ApiToken, error)
	List(ctx context.Context, userID uuid.UUID) ([]repository.ListApiTokensRow, error)
	Revoke(ctx context.Context, userID, id uuid.UUID) error
}

type apiTokenUsecase struct {
	repo *repository.Queries
}

func NewApiTokenUsecase(repo *repository.Queries) ApiTokenUsecase {
	return &apiTokenUsecase{repo: repo}
}

func (u *apiTokenUsecase) Create(ctx context.Context, userID uuid.UUID, name string) (string, *repository.ApiToken, error) {
	rawToken, tokenHash, err := auth.GeneratePAT()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	token, err := u.repo.CreateApiToken(ctx, repository.CreateApiTokenParams{
		UserID:    userID,
		Name:      name,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create api token: %w", err)
	}

	return rawToken, &token, nil
}

func (u *apiTokenUsecase) List(ctx context.Context, userID uuid.UUID) ([]repository.ListApiTokensRow, error) {
	tokens, err := u.repo.ListApiTokens(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list api tokens: %w", err)
	}
	return tokens, nil
}

func (u *apiTokenUsecase) Revoke(ctx context.Context, userID, id uuid.UUID) error {
	err := u.repo.DeleteApiToken(ctx, repository.DeleteApiTokenParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke api token: %w", err)
	}
	return nil
}
