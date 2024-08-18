package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"

	"adamolsen.dev/buggins/internal/inat"
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
	projectId := e.GetString("INATURALIST_PROJECT_ID", "")
	channelId := e.GetString("INATURALIST_CHANNEL_ID", "")
	cronPattern := e.GetString("INATURALIST_CRON_PATTERN", "0 * * * *")
	pageSize := e.GetInt("INATURALIST_FETCH_PAGE_SIZE", 10)

	if token == "" || guildID == "" || projectId == "" || channelId == "" {
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

	service := inat.NewService(inat.ServiceConfig{
		ProjectID: projectId,
		PageSize:  pageSize,
		Store:     s,
	})

	bot := inat.NewBot(inat.BotConfig{
		Service:     service,
		CronPattern: cronPattern,
		Discord:     discord,
		ChannelID:   channelId,
		GuildID:     guildID,
	})

	err = discord.Open()

	if err != nil {
		log.Fatal(err)
	}

	bot.StartPosting()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = discord.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
