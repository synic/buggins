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
	viper.SetDefault("DatabaseURL", "db.sqlite")
	viper.BindEnv("DatabaseURL", "DATABASE_URL")
	rootCmd.PersistentFlags().
		StringVar(&databaseURL, "database-url", "db.sqlite", "Database connection URL (can be set with $DATABASE_URL in env)")
	viper.BindPFlag("DatabaseURL", rootCmd.PersistentFlags().Lookup("database-url"))

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}
