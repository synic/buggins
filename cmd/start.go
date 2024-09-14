package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start buggins bot and connect to Discord",
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(getProviders(), fx.Invoke(func(bot) {})).Run()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
