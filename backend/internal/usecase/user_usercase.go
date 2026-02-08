package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gigaonion/taskalyst/backend/internal/infra/db"
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/pkg/auth"
	"github.com/gigaonion/taskalyst/backend/pkg/password"
)

type UserUsecase interface {
	SignUp(ctx context.Context, email, plainPassword, name string) (*repository.User, error)
	Login(ctx context.Context, email, plainPassword string) (*auth.TokenPair, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*repository.User, error)
}

type userUsecase struct {
	repo   *repository.Queries
	config *config.Config
	txManager db.TxManager
}

func NewUserUsecase(repo *repository.Queries, txManager db.TxManager, cfg *config.Config) UserUsecase {
	return &userUsecase{
		repo:   repo,
		config: cfg,
		txManager: txManager,
	}
}

func (u *userUsecase) SignUp(ctx context.Context, email, plainPassword, name string) (*repository.User, error) {
	// Password
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, err
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
			return err
		}


		user = createdUser
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("signup failed: %w", err)
	}

	return &user, nil
}

func (u *userUsecase) Login(ctx context.Context, email, plainPassword string) (*auth.TokenPair, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Password
	match, err := password.Check(plainPassword, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check password: %w", err)
	}
	if !match {
		return nil, fmt.Errorf("invalid email or password")
	}

	// トークンの生成
	tokens, err := auth.GenerateTokenPair(user.ID, string(user.Role), u.config.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}
func (u *userUsecase) GetProfile(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
