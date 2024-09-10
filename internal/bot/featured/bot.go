package featured

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"

	"github.com/synic/buggins/internal/store"
)

type BotConfig struct {
	ChannelID             string `env:"FEATURED_CHANNEL_ID, required"`
	RequiredReactionCount int    `env:"FEATURED_REACTION_COUNT, default=8"`
}

type Bot struct {
	discord *discordgo.Session
	db      *store.Queries
	BotConfig
	isStarted bool
}

func New(discord *discordgo.Session, db *store.Queries, config BotConfig) *Bot {
	return &Bot{BotConfig: config, discord: discord, db: db}
}

func InitFromEnv(discord *discordgo.Session, db *store.Queries) (*Bot, error) {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, fmt.Errorf("inatobs bot missing config: %w", err)
	}

	return New(discord, db, c), nil
}

func (b *Bot) Start() {
	if !b.isStarted {
		b.isStarted = true
		b.registerHandlers()
		log.Println("Started featured bot")
	}
}

func (b *Bot) registerHandlers() {
	b.discord.AddHandler(func(d *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if !b.isStarted {
			return
		}

		m, err := b.discord.ChannelMessage(r.ChannelID, r.MessageID)

		if err != nil {
			log.Printf("error fetching message ID `%s`: %v", r.MessageID, err)
		}

		imgCount := getImageAttachmentCount(m.Attachments)
		reactionCount := getReactionCount(m.Reactions)

		if imgCount > 0 && reactionCount >= b.RequiredReactionCount {
			isFeatured, err := b.db.IsMessageFeatured(
				context.Background(),
				store.IsMessageFeaturedParams{
					ChannelID: r.ChannelID,
					MessageID: r.MessageID,
				},
			)

			if err != nil {
				log.Printf(
					"couldn't determine if message is featured %s %s: %v",
					r.ChannelID,
					r.MessageID,
					err,
				)
				return
			}

			if isFeatured > 0 {
				log.Printf(
					"message is already featured, skipping %s %s",
					r.ChannelID,
					r.MessageID,
				)
				return
			}

			_, err = b.db.SaveFeaturedMessage(
				context.Background(),
				store.SaveFeaturedMessageParams{
					ChannelID: r.ChannelID,
					MessageID: r.MessageID,
				},
			)

			if err != nil {
				log.Printf(
					"couldn't save featured message to db %s %s: %v",
					r.ChannelID,
					r.MessageID,
					err,
				)
				return
			}

			files := make([]*discordgo.File, 0, len(m.Attachments))

			for _, a := range m.Attachments {
				if !strings.Contains(a.ContentType, "image") {
					continue
				}

				r, err := http.Get(a.URL)

				if err != nil {
					log.Printf("unable to retrieve data for photo `%s`: %v", a.URL, err)
					continue
				}

				defer r.Body.Close()
				files = append(files, &discordgo.File{
					Name:        a.Filename,
					ContentType: a.ContentType,
					Reader:      r.Body,
				})
			}

			b.discord.ChannelMessageSendComplex(
				b.ChannelID,
				&discordgo.MessageSend{
					Files: files,
					Embed: &discordgo.MessageEmbed{
						Author: &discordgo.MessageEmbedAuthor{
							Name:    m.Author.Username,
							IconURL: m.Author.AvatarURL(""),
						},
						URL: fmt.Sprintf(
							"https://discord.com/channels/@me/%s/%s",
							r.ChannelID,
							r.MessageID,
						),
						Title: fmt.Sprintf(
							"%s's post had enough reactions to be included in the Hall of Fame!",
							m.Author.Username,
						),
					},
				},
			)
		}

	})
}

func getImageAttachmentCount(attachments []*discordgo.MessageAttachment) int {
	if len(attachments) <= 1 {
		return 0
	}

	count := 0

	for _, attachment := range attachments {
		if strings.Contains(attachment.ContentType, "image") {
			count += 1
		}
	}

	return count
}

func getReactionCount(reactions []*discordgo.MessageReactions) int {
	count := 0

	for _, reaction := range reactions {
		if reaction.Me {
			continue
		}
		count += reaction.Count
	}

	return count
}
