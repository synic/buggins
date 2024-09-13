package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

type bot struct{ modules []m.Module }
type botParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Discord   *discordgo.Session
	Modules   []m.Module `group:"modules"`
}

func newDiscordSession(conf conf.Config) (*discordgo.Session, error) {
	return discordgo.New(fmt.Sprintf("Bot %s", conf.DiscordToken))
}

func newBot(params botParams) bot {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
				log.Printf("User '%s' connected to discord!", r.User.Username)

				for _, module := range params.Modules {
					module.Start()
				}
			})

			if err := params.Discord.Open(); err != nil {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := params.Discord.Close(); err != nil {
				return err
			}

			return nil
		},
	})

	return bot{modules: params.Modules}
}

func newDatabase(conf conf.Config) (*store.Queries, error) {
	return store.Init(conf.DatabaseURL)
}
