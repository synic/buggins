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
	db          *store.Queries
	options     Options
	isStarted   bool
	optionsLock sync.RWMutex
}

func New(db *store.Queries) (*Module, error) {
	options, err := fetchModuleOptions(db)

	if err != nil {
		return nil, fmt.Errorf("unable to parse featured options: %w", err)
	}

	return &Module{options: options, db: db}, nil
}

func (m *Module) Start(discord *discordgo.Session) error {
	if !m.isStarted {
		m.isStarted = true
		m.registerHandlers(discord)
		log.Println("started featured module")
		log.Printf(" -> guilds: %+v", m.Options().Guilds)
	}
	return nil
}

func (m *Module) Name() string {
	return moduleName
}

func (m *Module) Options() Options {
	m.optionsLock.RLock()
	defer m.optionsLock.RUnlock()
	return m.options
}

func (m *Module) SetOptions(options Options) {
	m.optionsLock.Lock()
	defer m.optionsLock.Unlock()
	m.options = options
}

func (m *Module) ReloadConfig(discord *discordgo.Session, db *store.Queries) error {
	options, err := fetchModuleOptions(db)

	if err != nil {
		return fmt.Errorf("unable to parse featured options: %w", err)
	}

	m.SetOptions(options)
	log.Printf(" -> guilds: %+v", m.Options().Guilds)
	return nil
}

func (m *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range m.Options().Guilds {
		if o.ID == guildID {
			return o, nil
		}
	}

	return GuildOptions{}, errors.New("guild not configured")
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	discord.AddHandler(func(d *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if !m.isStarted {
			return
		}

		options, err := m.getGuildOptions(r.GuildID)

		if err != nil || options.ChannelID != r.ChannelID {
			return
		}

		msg, err := d.ChannelMessage(r.ChannelID, r.MessageID)

		if err != nil {
			log.Printf("error fetching message ID `%s`: %v", r.MessageID, err)
			return
		}

		hasEnough := hasEnoughReactions(msg.Reactions, options.RequiredReactionCount)
		imgCount := getImageAttachmentCount(msg.Attachments)

		if imgCount > 0 && hasEnough {
			isFeatured, err := m.db.FindIsMessageFeatured(
				context.Background(),
				store.FindIsMessageFeaturedParams{
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

			discord.ChannelMessageSendComplex(
				options.ChannelID,
				&discordgo.MessageSend{
					Content: fmt.Sprintf(
						":partying_face: Congratulations, <@%s>, your [post](https://discord.com/channels/@me/%s/%s) made the Hall of Fame!",
						msg.Author.ID,
						r.ChannelID,
						r.MessageID,
					),
					Files: files,
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

func hasEnoughReactions(
	reactions []*discordgo.MessageReactions,
	enough int,
) bool {
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
