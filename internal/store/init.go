package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(dsn string) (*Queries, error) {
	pool, err := pgxpool.New(context.Background(), dsn)

	if err != nil {
		return nil, err
	}

	if runMigrations {
		if err := RunMigrations(pool); err != nil {
			pool.Close()
			return nil, err
		}
	}

	db := New(pool)
	return db, nil
}

func Migrate(dsn string) error {
	pool, err := pgxpool.New(context.Background(), dsn)

	if err != nil {
		return err
	}

	defer pool.Close()
	return RunMigrations(pool)
}
