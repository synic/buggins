#!/bin/sh

export GOOSE_MIGRATION_DIR=./internal/store/migrations
export GOOSE_DRIVER=sqlite3
export GOOSE_DBSTRING=./db.sqlite

NAME="$@"

if [ "$NAME" = "" ]; then
  read -p "Enter a name for the migration: " NAME
fi

goose create $NAME sql
