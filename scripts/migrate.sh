#!/bin/sh

export GOOSE_MIGRATION_DIR=./internal/store/migrations
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DATABASE_URL"

CMD="$@"

if [ "$CMD" = "" ]; then
  CMD="up"
fi

goose $CMD
