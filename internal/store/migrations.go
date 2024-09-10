package store

import (
	"database/sql"
	"embed"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var EmbeddedMigrations embed.FS
var runMigrations = false

func MaybeRunMigrations(dialect string, db *sql.DB) {
	if !runMigrations {
		return
	}

	log.Println("Running embedded migrations")

	goose.SetBaseFS(EmbeddedMigrations)

	if err := goose.SetDialect(dialect); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
