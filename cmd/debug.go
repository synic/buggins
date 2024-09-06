//go:build debug

package cmd

import (
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Overload(".env")
}
