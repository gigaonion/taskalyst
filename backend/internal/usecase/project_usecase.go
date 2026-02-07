package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
)

type ProjectUsecase interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, name string, rootType repository.RootCategoryType, color string) (*repository.Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]repository.Category, error)
	CreateProject(ctx context.Context, userID, categoryID uuid.UUID, title, description string) (*repository.Project, error)
	ListProjects(ctx context.Context, userID uuid.UUID, isArchived *bool) ([]repository.ListProjectsRow, error)
}

type projectUsecase struct {
	repo *repository.Queries
}

func NewProjectUsecase(repo *repository.Queries) ProjectUsecase {
	return &projectUsecase{repo: repo}
}

// --- Helper Functions ---

// toText converts *string to pgtype.Text
func toText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// toTextFromStr converts string to pgtype.Text (empty string check optional)
func toTextFromStr(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// --- Category ---

func (u *projectUsecase) CreateCategory(ctx context.Context, userID uuid.UUID, name string, rootType repository.RootCategoryType, color string) (*repository.Category, error) {
	arg := repository.CreateCategoryParams{
		UserID:   userID,
		Name:     name,
		RootType: rootType,
		// color は必須ではないが、ハンドラからは空文字または有効な値が来る想定
		Color:    toTextFromStr(color), 
	}
	
	category, err := u.repo.CreateCategory(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return &category, nil
}

func (u *projectUsecase) ListCategories(ctx context.Context, userID uuid.UUID) ([]repository.Category, error) {
	categories, err := u.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// --- Project ---

func (u *projectUsecase) CreateProject(ctx context.Context, userID, categoryID uuid.UUID, title, description string) (*repository.Project, error) {
	arg := repository.CreateProjectParams{
		UserID:      userID,
		CategoryID:  categoryID,
		Title:       title,
		Description: toTextFromStr(description), // string -> pgtype.Text
		IsArchived:  false,
	}

	project, err := u.repo.CreateProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return &project, nil
}

func (u *projectUsecase) ListProjects(ctx context.Context, userID uuid.UUID, isArchived *bool) ([]repository.ListProjectsRow, error) {
	// isArchived (nullable boolean) の処理
	var argIsArchived pgtype.Bool
	if isArchived != nil {
		argIsArchived = pgtype.Bool{Bool: *isArchived, Valid: true}
	} else {
		argIsArchived = pgtype.Bool{Valid: false}
	}

	arg := repository.ListProjectsParams{
		UserID:     userID,
		IsArchived: argIsArchived, // *bool -> pgtype.Bool
	}
	
	projects, err := u.repo.ListProjects(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}
