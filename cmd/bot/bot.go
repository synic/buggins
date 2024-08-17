package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"

	"adamolsen.dev/buggins/internal/db"
	"adamolsen.dev/buggins/internal/inat"
	"adamolsen.dev/buggins/internal/pkg/env"
)

func Start(db *db.Queries) {
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
		DB:        db,
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
