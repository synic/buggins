package store

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init() (*Queries, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

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

func Migrate() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		return err
	}

	defer pool.Close()
	return RunMigrations(pool)
}
