package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %V, rb err: %v", err, rbErr)
		}

		return err
	}
	return tx.Commit(ctx)
}
