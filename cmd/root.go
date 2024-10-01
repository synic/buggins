package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	discordToken string
	databaseURL  string
)

var rootCmd = &cobra.Command{
	Use:   "buggins",
	Short: "Discord bot for the Macromania server",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.SetDefault("DatabaseFile", "db.sqlite")
	viper.BindEnv("DatabaseFile", "DATABASE_FILE")
	rootCmd.PersistentFlags().
		StringVar(&databaseURL, "database-file", "db.sqlite", "Database connection URL (can be set with $DATABASE_URL in env)")
	viper.BindPFlag("DatabaseFile", rootCmd.PersistentFlags().Lookup("database-url"))

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}
