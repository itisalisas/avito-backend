package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type BaseRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewBaseRepository(db *sql.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	r.tx = tx
	return tx, nil
}

func (r *BaseRepository) Commit() error {
	if r.tx == nil {
		return fmt.Errorf("no active transaction to commit")
	}
	if err := r.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	r.tx = nil
	return nil
}

func (r *BaseRepository) Rollback() error {
	if r.tx == nil {
		return nil
	}
	if err := r.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	r.tx = nil
	return nil
}
