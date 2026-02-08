package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/pkg/auth"
	"github.com/gigaonion/taskalyst/backend/pkg/password"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserUsecase interface {
	SignUp(ctx context.Context, email, plainPassword, name string) (*repository.User, error)
	Login(ctx context.Context, email, plainPassword string) (*auth.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*repository.User, error)
}

type userUsecase struct {
	repo      *repository.Queries
	config    *config.Config
	txManager db.TxManager
}

func NewUserUsecase(repo *repository.Queries, txManager db.TxManager, cfg *config.Config) UserUsecase {
	return &userUsecase{
		repo:      repo,
		config:    cfg,
		txManager: txManager,
	}
}

func (u *userUsecase) SignUp(ctx context.Context, email, plainPassword, name string) (*repository.User, error) {
	// Password
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user repository.User

	err = u.txManager.ReadCommitted(ctx, func(q *repository.Queries) error {
		// User
		arg := repository.CreateUserParams{
			Email:        email,
			PasswordHash: hash,
			Name:         name,
			Role:         repository.UserRoleUSER,
		}

		createdUser, err := q.CreateUser(ctx, arg)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" { // unique_violation
					return NewConflictError("email already exists")
				}
			}
			return err
		}

		user = createdUser
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *userUsecase) Login(ctx context.Context, email, plainPassword string) (*auth.TokenPair, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewUnauthorizedError("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Password
	match, err := password.Check(plainPassword, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check password: %w", err)
	}
	if !match {
		return nil, NewUnauthorizedError("invalid email or password")
	}

	// トークンの生成
	tokens, err := auth.GenerateTokenPair(user.ID, string(user.Role), u.config.JWTSecret, u.config.JWTRefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}

func (u *userUsecase) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// リフレッシュトークンの検証
	claims, err := auth.ValidateToken(refreshToken, u.config.JWTRefreshSecret)
	if err != nil {
		return nil, NewUnauthorizedError("invalid refresh token")
	}

	// ユーザーの存在確認
	user, err := u.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewUnauthorizedError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 新しいトークンペアの生成
	tokens, err := auth.GenerateTokenPair(user.ID, string(user.Role), u.config.JWTSecret, u.config.JWTRefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}

func (u *userUsecase) GetProfile(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
