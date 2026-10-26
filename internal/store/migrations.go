package store

import (
	"embed"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var EmbeddedMigrations embed.FS
var runMigrations = false

func RunMigrations(pool *pgxpool.Pool) error {
	log.Println("Running embedded migrations")

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
