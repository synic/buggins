package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"

	"adamolsen.dev/buggins/internal/discord/inat"
	"adamolsen.dev/buggins/internal/discord/thisthat"
	"adamolsen.dev/buggins/internal/pkg/env"
	"adamolsen.dev/buggins/internal/store"
)

func main() {
	conn, err := sql.Open("sqlite3", "./data/database.sqlite")

	if err != nil {
		log.Fatal(err)
	}

	store.RunMigrations("sqlite3", conn)

	s := store.New(conn)
	e := env.New()

	token := e.GetString("DISCORD_TOKEN", "")
	guildID := e.GetString("DISCORD_GUILD_ID", "")
	inatProjectID := e.GetString("INATURALIST_PROJECT_ID", "")
	inatChannelID := e.GetString("INATURALIST_CHANNEL_ID", "")
	inatCronPattern := e.GetString("INATURALIST_CRON_PATTERN", "0 * * * *")
	inatPageSize := e.GetInt("INATURALIST_FETCH_PAGE_SIZE", 10)
	thisThatChannelID := e.GetString("THISTHAT_CHANNEL_ID", "")

	if token == "" || guildID == "" || inatProjectID == "" || inatChannelID == "" {
		log.Fatal(
			"You must set the DISCORD_TOKEN, DISCORD_GUILD_ID, " +
				"INATURALIST_CHANNEL_ID, and INATURALIST_PROJECT_ID " +
				"environment variables.",
		)
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", token))

	if err != nil {
		log.Fatal(err)
	}

	inatBot := inat.New(inat.BotConfig{
		CronPattern: inatCronPattern,
		Discord:     discord,
		ChannelID:   inatChannelID,
		GuildID:     guildID,
		ProjectID:   inatProjectID,
		PageSize:    inatPageSize,
		Store:       s,
	})

	thisthatBot := thisthat.New(
		thisthat.BotConfig{
			Discord:   discord,
			ChannelID: thisThatChannelID,
		},
	)

	discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
		log.Printf("User %s connected to discord!", r.User.Username)

		inatBot.Start()

		if thisThatChannelID != "" {
			thisthatBot.Start()
		}
	})

	err = discord.Open()

	if err != nil {
		log.Fatal(err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = discord.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
