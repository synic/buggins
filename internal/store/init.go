package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Init(fileLocation string) (*Queries, error) {
	conn, err := sql.Open("sqlite3", fileLocation)

	if err != nil {
		return nil, err
	}

	MaybeRunMigrations("sqlite3", conn)
	db := New(conn)
	return db, nil
}
