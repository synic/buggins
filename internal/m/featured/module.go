package featured

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

type GuildOptions struct {
	Name                  string `mapstructure:"name"`
	ID                    string `mapstructure:"id"`
	ChannelID             string `mapstructure:"channel_id"`
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
	var options Options
	err := c.Populate("featured", &options)

	if err != nil {
		return providerResult{}, err
	}

	return providerResult{Module: New(discord, db, options)}, nil
}

type Module struct {
	discord   *discordgo.Session
	db        *store.Queries
	options   Options
	isStarted bool
}

func New(discord *discordgo.Session, db *store.Queries, options Options) *Module {
	return &Module{options: options, discord: discord, db: db}
}

func (m *Module) Start() {
	if !m.isStarted {
		m.isStarted = true
		m.registerHandlers()
		log.Println("started featured module")
	}
}

func (m *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range m.options.Guilds {
		if o.ID == guildID {
			return o, nil
		}
	}

	return GuildOptions{}, errors.New("guild not configured")
}

func (m *Module) registerHandlers() {
	m.discord.AddHandler(func(d *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if !m.isStarted {
			return
		}

		options, err := m.getGuildOptions(r.GuildID)

		if err != nil || options.ChannelID != r.ChannelID {
			return
		}

		msg, err := m.discord.ChannelMessage(r.ChannelID, r.MessageID)

		if err != nil {
			log.Printf("error fetching message ID `%s`: %v", r.MessageID, err)
			return
		}

		reaction := getWinningReaction(msg.Reactions, options.RequiredReactionCount)
		imgCount := getImageAttachmentCount(msg.Attachments)

		if imgCount > 0 && reaction != nil {
			isFeatured, err := m.db.IsMessageFeatured(
				context.Background(),
				store.IsMessageFeaturedParams{
					ChannelID: r.ChannelID,
					MessageID: r.MessageID,
					GuildID:   r.GuildID,
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

			_, err = m.db.SaveFeaturedMessage(
				context.Background(),
				store.SaveFeaturedMessageParams{
					ChannelID: r.ChannelID,
					MessageID: r.MessageID,
					GuildID:   r.GuildID,
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

			files := make([]*discordgo.File, 0, len(msg.Attachments))

			for _, a := range msg.Attachments {
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

			m.discord.ChannelMessageSendComplex(
				options.ChannelID,
				&discordgo.MessageSend{
					Files: files,
					Embed: &discordgo.MessageEmbed{
						Author: &discordgo.MessageEmbedAuthor{
							Name:    msg.Author.Username,
							IconURL: msg.Author.AvatarURL(""),
						},
						URL: fmt.Sprintf(
							"https://discord.com/channels/@me/%s/%s",
							r.ChannelID,
							r.MessageID,
						),
						Title: fmt.Sprintf(
							"%s's post had enough reactions to be included in the Hall of Fame!",
							msg.Author.Username,
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

func getWinningReaction(reactions []*discordgo.MessageReactions, enough int) *discordgo.Emoji {
	if enough == 0 {
		enough = 6
	}

	for _, reaction := range reactions {
		if reaction.Me {
			continue
		}

		if reaction.Count >= enough {
			return reaction.Emoji
		}
	}

	return nil
}
