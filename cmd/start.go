package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var (
	ipcSocket             string
	shouldStartIpcService bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start buggins bot and connect to Discord",
	Run: func(cmd *cobra.Command, args []string) {
		bind := ""
		if ipcSocket != "" && shouldStartIpcService {
			bind = ipcSocket
		}

		ipcService := provideIpcService(bind)
		fx.New(
			getProviders(viper.GetString("DatabaseFile")),
			fx.Provide(newDiscordSession(viper.GetString("DiscordToken"))),
			fx.Invoke(func(*discordgo.Session) {}),
			ipcService,
		).
			Run()
	},
}

func init() {
	viper.BindEnv("DiscordToken", "DISCORD_TOKEN")

	startCmd.Flags().StringVar(&ipcSocket, "ipc-socket", "/tmp/buggins-ipc.sock", "IPC bind")
	startCmd.Flags().BoolVar(&shouldStartIpcService, "start-ipc", false, "Start IPC Server")
	startCmd.Flags().
		StringVar(&discordToken, "discord-token", "", "Discord token (can be set with $DISCORD_TOKEN in env)")
	viper.BindPFlag("DiscordToken", startCmd.Flags().Lookup("discord-token"))
	startCmd.MarkFlagRequired("discord-token")

	if viper.IsSet("DiscordToken") {
		startCmd.Flags().
			SetAnnotation("discord-token", cobra.BashCompOneRequiredFlag, []string{"false"})
	}

	rootCmd.AddCommand(startCmd)
}
