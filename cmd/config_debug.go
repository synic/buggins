//go:build debug

package cmd

import (
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	logger.SetLevel(log.DebugLevel)
}
