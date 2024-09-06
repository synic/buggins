package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/cobra"

	"github.com/synic/buggins/internal/store"
)

func startBot() {
	var (
		conf config
		bots []bot = make([]bot, 0, 3)
	)

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
			log.Printf("error starting bot: %v", err)
			continue
		}

		bots = append(bots, bot)
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

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the bot",
	Long:  "Start the bot and connect to discord",
	Run: func(cmd *cobra.Command, args []string) {
		startBot()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
