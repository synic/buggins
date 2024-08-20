package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sethvargo/go-envconfig"

	"adamolsen.dev/buggins/internal/bot/inatlookup"
	"adamolsen.dev/buggins/internal/bot/inatobs"
	"adamolsen.dev/buggins/internal/bot/thisthat"
	"adamolsen.dev/buggins/internal/store"
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
}

func main() {
	var (
		conf config
		bots = make([]bot, 0, 3)
	)

	godotenv.Load()

	if err := envconfig.Process(context.Background(), &conf); err != nil {
		log.Fatal(err)
	}

	db, err := store.Init(conf.DatabaseURL)

	if err != nil {
		log.Fatal(err)
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", conf.DiscordToken))

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range initFuncs {
		bot, err := f(discord, db)

		if err != nil {
			log.Print(err)
			continue
		}

		bots = append(bots, bot)
	}

	if len(bots) <= 0 {
		log.Fatal("no bots to start, bailing")
	}

	discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
		log.Printf("User '%s' connected to discord!", r.User.Username)

		for _, bot := range bots {
			bot.Start()
		}
	})

	if err := discord.Open(); err != nil {
		log.Fatal(err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	if err := discord.Close(); err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
