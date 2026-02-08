package db

import (
	"context"
	"fmt"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (tm *txManager) ReadCommitted(ctx context.Context, fn func(q *repository.Queries) error) (err error) {
	tx, txErr := tm.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if txErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", txErr)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// fn実行用に新しいQueriesインスタンスを作成
	q := repository.New(tx)

	if err = fn(q); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
