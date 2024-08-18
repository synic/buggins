#!/bin/sh

export GOOSE_MIGRATION_DIR=./internal/store/migrations
export GOOSE_DRIVER=sqlite3
export GOOSE_DBSTRING=./data/database.sqlite

CMD="$@"

if [ "$CMD" = "" ]; then
  CMD="up"
fi

goose $CMD