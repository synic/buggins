package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "buggins",
	Short: "Discord bot for the Macromania server",
	Long:  "Discord bot for the Macromania server",
	Run: func(cmd *cobra.Command, args []string) {
		startBot()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}
