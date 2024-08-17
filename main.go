package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"adamolsen.dev/buggins/cmd/bot"
	botdb "adamolsen.dev/buggins/internal/db"
)

func main() {
	db, err := sql.Open("sqlite", "./data/database.sqlite")

	if err != nil {
		log.Fatal(err)
	}

	queries := botdb.New(db)

	bot.Start(queries)
}
