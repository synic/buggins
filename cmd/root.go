package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	viper.BindEnv("DiscordToken", "DISCORD_TOKEN")
	rootCmd.PersistentFlags().
		String("database-url", "", "Database connection URL (can be set with $DATABASE_URL in env)")
	rootCmd.PersistentFlags().
		String("discord-token", "", "Discord token (can be set with $DISCORD_TOKEN in env)")
	viper.BindPFlag("DatabaseURL", rootCmd.PersistentFlags().Lookup("database-url"))
	viper.BindPFlag("DiscordToken", rootCmd.Flags().Lookup("discord-token"))

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}
