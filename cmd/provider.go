package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/m/featured"
	"github.com/synic/buggins/internal/m/inatlookup"
	"github.com/synic/buggins/internal/m/inatobs"
	"github.com/synic/buggins/internal/m/thisthat"
	"github.com/synic/buggins/internal/store"
)

type bot struct{ modules []m.Module }
type botParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Discord   *discordgo.Session
	Modules   []m.Module `group:"modules"`
}

func getProviders() fx.Option {
	return fx.Options(
		fx.Provide(newDiscordSession),
		fx.Provide(newDatabase),
		fx.Provide(featured.Provider),
		fx.Provide(inatobs.Provider),
		fx.Provide(inatlookup.Provider),
		fx.Provide(thisthat.Provider),
		fx.Provide(newBot),
	)
}

func newDiscordSession() (*discordgo.Session, error) {
	token := viper.GetString("DiscordToken")
	if token == "" {
		log.Fatalln("Discord token not set. Pass --discord-token or set $DISCORD_TOKEN")
	}
	return discordgo.New(fmt.Sprintf("Bot %s", token))
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

func newDatabase() (*store.Queries, error) {
	return store.Init(viper.GetString("DatabaseURL"))
}
