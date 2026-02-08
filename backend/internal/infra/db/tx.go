package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
)

// トランザクション制御
type TxManager interface {
	ReadCommitted(ctx context.Context, fn func(q *repository.Queries) error) error
}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) TxManager {
	return &txManager{pool: pool}
}

func (tm *txManager) ReadCommitted(ctx context.Context, fn func(q *repository.Queries) error) error {
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// fn実行用に新しいQueriesインスタンスを作成
	q := repository.New(tx)
	
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
