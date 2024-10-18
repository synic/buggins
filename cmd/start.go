package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var (
	discordToken          string
	ipcSocket             string
	shouldStartIpcService = false
)

var startCmd = &cli.Command{
	Name:  "start",
	Usage: "Start buggins bot and connect to Discord",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "discord-token",
			Required:    true,
			Destination: &discordToken,
			Usage:       "Discord token",
			EnvVars:     []string{"DISCORD_TOKEN"},
		},
		&cli.BoolFlag{
			Name:        "start-ipc",
			Value:       false,
			Destination: &shouldStartIpcService,
			Usage:       "Start IPC server",
		},
		&cli.StringFlag{
			Name:        "ipc-socket",
			Value:       "/tmp/buggins-ipc.sock",
			Usage:       "IPC bind socket",
			Destination: &ipcSocket,
		},
	},
	Action: func(ctx *cli.Context) error {
		bind := ""
		if ipcSocket != "" && shouldStartIpcService {
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
	},
}

func init() {
	app.Commands = append(app.Commands, startCmd)
}
