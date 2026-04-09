package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/synic/glap"
	"go.uber.org/fx"
)

func init() {
	cmd := glap.NewCommand("start").
		About("Start buggins bot and connect to Discord").
		Arg(glap.NewArg("discord-token").
			Required(true).
			Env("DISCORD_TOKEN").
			Help("Discord token")).
		Arg(glap.NewArg("start-ipc").
			Action(glap.SetTrue).
			Help("Start IPC server")).
		Arg(glap.NewArg("ipc-socket").
			Default("/tmp/buggins-ipc.sock").
			Help("IPC bind socket")).
		Run(func(m *glap.Matches) error {
			discordToken, _ := m.GetString("discord-token")
			shouldStartIpc, _ := m.GetBool("start-ipc")
			ipcSocket, _ := m.GetString("ipc-socket")

			bind := ""
			if ipcSocket != "" && shouldStartIpc {
				bind = ipcSocket
			}

			ipcService := provideIpcService(bind)
			fx.New(
				providers(databaseFile),
				fx.Provide(newDiscordSession(discordToken)),
				fx.Invoke(func(*discordgo.Session) {}),
				ipcService,
			).
				Run()
			return nil
		})

	RegisterCommand(cmd)
}
