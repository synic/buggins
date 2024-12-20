package featured

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "featured"
)

type Module struct {
	db         *store.Queries
	logger     *slog.Logger
	config     []GuildConfig
	configLock sync.RWMutex
}

func New(db *store.Queries, logger *slog.Logger) (*Module, error) {
	return &Module{db: db, logger: logger}, nil
}

func (m *Module) Start(ctx context.Context, discord *discordgo.Session, db *store.Queries) error {
	config, err := mod.FetchModuleConfiguration[GuildConfig](ctx, db, moduleName)

	if err != nil {
		return err
	}

	m.SetConfig(config)
	m.registerHandlers(discord)
	m.logger.Info("started module")
	m.logger.Info(" -> config", "guilds", m.Config())
	return nil
}

func Provider(db *store.Queries, logger *slog.Logger) (mod.ModuleProviderResult, error) {
	module, err := New(db, logger.With("mod", moduleName))

	if err != nil {
		return mod.ModuleProviderResult{}, err
	}

	return mod.ModuleProviderResult{Module: module}, nil
}

func (m *Module) Name() string {
	return moduleName
}

func (m *Module) Config() []GuildConfig {
	m.configLock.RLock()
	defer m.configLock.RUnlock()
	return m.config
}

func (m *Module) SetConfig(config []GuildConfig) {
	m.configLock.Lock()
	defer m.configLock.Unlock()
	m.config = config
}

func (m *Module) ReloadConfig(
	ctx context.Context,
	discord *discordgo.Session,
	db *store.Queries,
) error {
	config, err := mod.FetchModuleConfiguration[GuildConfig](ctx, db, moduleName)

	if err != nil {
		return fmt.Errorf("unable to parse featured options: %w", err)
	}

	m.SetConfig(config)
	m.logger.Info(" -> config", "guilds", m.Config())
	return nil
}

func (m *Module) guildConfig(guildID string) (GuildConfig, error) {
	for _, o := range m.Config() {
		if o.ID == guildID {
			return o, nil
		}
	}

	return GuildConfig{}, errors.New("guild not configured")
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	discord.AddHandler(func(d *discordgo.Session, r *discordgo.MessageReactionAdd) {
		config, err := m.guildConfig(r.GuildID)

		if err != nil {
			return
		}

		msg, err := d.ChannelMessage(r.ChannelID, r.MessageID)

		if err != nil {
			m.logger.Info("error fetching message ID", "message", r.MessageID, "err", err)
			return
		}

		reactionCount := starReactionCount(msg.Reactions)
		imgCount := imageAttachmentCount(msg.Attachments)

		if imgCount > 0 && reactionCount >= config.RequiredReactionCount {
			isFeatured, err := m.db.FindIsMessageFeatured(
				context.Background(),
				store.FindIsMessageFeaturedParams{
					ChannelID: r.ChannelID,
					MessageID: r.MessageID,
					GuildID:   r.GuildID,
				},
			)

			if err != nil {
				m.logger.Warn(
					"couldn't determine if message is featured",
					"channel",
					r.ChannelID,
					"message",
					r.MessageID,
					"err",
					err,
				)
				return
			}

			if isFeatured > 0 {
				m.logger.Warn(
					"message is already featured, skipping",
					"channel",
					r.ChannelID,
					"message",
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
				m.logger.Warn(
					"couldn't save featured message to db",
					"channel",
					r.ChannelID,
					"message",
					r.MessageID,
					"err",
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
					m.logger.Error("unable to retrieve data for photo", "url", a.URL, "err", err)
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
				config.ChannelID,
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

func imageAttachmentCount(attachments []*discordgo.MessageAttachment) int {
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

func starReactionCount(
	reactions []*discordgo.MessageReactions,
) int {
	count := 0

	for _, reaction := range reactions {
		if reaction.Me {
			continue
		}

		if reaction.Emoji.Name == "⭐" {
			count += reaction.Count
		}
	}

	return count
}
