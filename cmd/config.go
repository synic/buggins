package cmd

import (
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"

	"github.com/synic/buggins/internal/bot/featured"
	"github.com/synic/buggins/internal/bot/inatlookup"
	"github.com/synic/buggins/internal/bot/inatobs"
	"github.com/synic/buggins/internal/bot/thisthat"
	"github.com/synic/buggins/internal/store"
)

type config struct {
	DiscordToken string `env:"DISCORD_TOKEN, required"`
	DatabaseURL  string `env:"DATABASE_URL, default=./data/database.sqlite"`
}

type bot interface{ Start() }
type botInitFunc = func(d *discordgo.Session, s *store.Queries) (bot, error)

var initFuncs = []botInitFunc{
	func(d *discordgo.Session, s *store.Queries) (bot, error) {
		return inatobs.InitFromEnv(d, s)
	},
	func(d *discordgo.Session, s *store.Queries) (bot, error) {
		return inatlookup.InitFromEnv(d)
	},
	func(d *discordgo.Session, s *store.Queries) (bot, error) {
		return thisthat.InitFromEnv(d)
	},
	func(d *discordgo.Session, s *store.Queries) (bot, error) {
		return featured.InitFromEnv(d, s)
	},
}
