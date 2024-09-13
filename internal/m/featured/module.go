package featured

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

type GuildOptions struct {
	Name                  string `mapstructure:"name"`
	ID                    int    `mapstructure:"id"`
	ChannelID             int    `mapstructure:"channel_id"`
	RequiredReactionCount int    `mapstructure:"reaction_count"`
}

type Options struct {
	Guilds []GuildOptions `mapstructure:"guilds"`
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func Provider(
	c conf.Config,
	discord *discordgo.Session,
	db *store.Queries,
) (providerResult, error) {
	options, err := conf.GetConfig[Options](c, "featured")

	if err != nil {
		return providerResult{}, err
	}

	return providerResult{Module: New(discord, db, options)}, nil
}

type Module struct {
	Options
	discord   *discordgo.Session
	db        *store.Queries
	isStarted bool
}

func New(discord *discordgo.Session, db *store.Queries, config Options) *Module {
	return &Module{Options: config, discord: discord, db: db}
}

func (b *Module) Start() {
	if !b.isStarted {
		b.isStarted = true
		b.registerHandlers()
		log.Println("Started featured module")
	}
}

func (b *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range b.Guilds {
		if strconv.Itoa(o.ID) == guildID {
			return o, nil
		}
	}

	return GuildOptions{}, errors.New("guild not configured")
}

func (b *Module) registerHandlers() {
	b.discord.AddHandler(func(d *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if !b.isStarted {
			return
		}

		options, err := b.getGuildOptions(r.GuildID)

		if err != nil || strconv.Itoa(options.ChannelID) != r.ChannelID {
			return
		}

		m, err := b.discord.ChannelMessage(r.ChannelID, r.MessageID)

		if err != nil {
			log.Printf("error fetching message ID `%s`: %v", r.MessageID, err)
			return
		}

		shouldBeFeatured := getWinningReaction(m.Reactions, options.RequiredReactionCount)
		imgCount := getImageAttachmentCount(m.Attachments)

		if imgCount > 0 && shouldBeFeatured {
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
				strconv.Itoa(options.ChannelID),
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
	if len(attachments) < 1 {
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

func getWinningReaction(reactions []*discordgo.MessageReactions, enough int) bool {
	if enough == 0 {
		enough = 6
	}

	for _, reaction := range reactions {
		if reaction.Me {
			continue
		}

		if reaction.Count >= enough {
			return true
		}
	}

	return false
}
