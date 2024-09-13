package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Init(url string) (*Queries, error) {
	conn, err := sql.Open("sqlite3", url)

	if err != nil {
		return nil, err
	}

	MaybeRunMigrations("sqlite3", conn)
	db := New(conn)
	return db, nil
}
