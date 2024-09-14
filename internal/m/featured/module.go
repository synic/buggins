package featured

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/synic/buggins/internal/store"
)

type Module struct {
	discord   *discordgo.Session
	db        *store.Queries
	options   Options
	isStarted bool
	mu        sync.Mutex
}

func New(discord *discordgo.Session, db *store.Queries) (*Module, error) {
	options, err := getModuleOptions(db)

	if err != nil {
		return nil, fmt.Errorf("unable to parse featured options: %w", err)
	}

	return &Module{options: options, discord: discord, db: db}, nil
}

func (m *Module) Start() {
	if !m.isStarted {
		m.isStarted = true
		m.registerHandlers()
		log.Println("started featured module")
		log.Printf(" -> guilds: %+v", m.GetOptions().Guilds)
	}
}

func (m *Module) GetName() string {
	return moduleName
}

func (m *Module) GetOptions() Options {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.options
}

func (m *Module) ReloadConfig(db *store.Queries) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	options, err := getModuleOptions(db)

	if err != nil {
		return fmt.Errorf("unable to parse featured options: %w", err)
	}

	m.options = options
	return nil
}

func (m *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range m.GetOptions().Guilds {
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
