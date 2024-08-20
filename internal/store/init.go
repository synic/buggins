package store

import (
	"database/sql"
)

func Init(url string) (*Queries, error) {
	conn, err := sql.Open("sqlite3", url)

	if err != nil {
		return nil, err
	}

	RunMigrations("sqlite3", conn)
	return New(conn), nil
}
