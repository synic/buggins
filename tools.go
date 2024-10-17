//go:build tools

package main

import (
	_ "github.com/air-verse/air"
	_ "github.com/pressly/goose/v3/cmd/goose"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
