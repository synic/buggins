package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m/featured"
	"github.com/synic/buggins/internal/m/inatlookup"
	"github.com/synic/buggins/internal/m/inatobs"
	"github.com/synic/buggins/internal/m/thisthat"
)

func getProviders(configFile string) fx.Option {
	return fx.Options(
		fx.Provide(conf.ProvideFromFile(configFile)),
		fx.Provide(newDiscordSession),
		fx.Provide(newDatabase),
		fx.Provide(featured.Provider),
		fx.Provide(inatobs.ProviderFromEnv),
		fx.Provide(inatlookup.ProviderFromEnv),
		fx.Provide(thisthat.ProviderFromEnv),
		fx.Provide(newBot),
	)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start buggins bot",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		fx.New(
			getProviders(configFile),
			fx.Invoke(func(bot) {}),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("config", "c", "", "configuration file location")
	startCmd.MarkFlagRequired("config")
}
