#!/bin/sh

export GOOSE_MIGRATION_DIR=./internal/store/migrations
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DATABASE_URL"

NAME="$@"

if [ "$NAME" = "" ]; then
  read -p "Enter a name for the migration: " NAME
fi

goose create $NAME sql
