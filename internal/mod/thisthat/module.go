package thisthat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"

	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "thisthat"
	emojis     = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
)

type Module struct {
	logger     *log.Logger
	config     []ChannelConfig
	configLock sync.RWMutex
}

func Provider(logger *log.Logger) (mod.ModuleProviderResult, error) {
	module, err := New(logger.With("mod", moduleName))

	if err != nil {
		return mod.ModuleProviderResult{}, err
	}

	return mod.ModuleProviderResult{Module: module}, nil
}

func New(logger *log.Logger) (*Module, error) {
	return &Module{logger: logger}, nil
}

func (m *Module) Name() string {
	return moduleName
}

func (m *Module) Config() []ChannelConfig {
	m.configLock.RLock()
	defer m.configLock.RUnlock()
	return m.config
}

func (m *Module) SetConfig(config []ChannelConfig) {
	m.configLock.Lock()
	defer m.configLock.Unlock()
	m.config = config
}

func (m *Module) ReloadConfig(
	ctx context.Context,
	discord *discordgo.Session,
	db *store.Queries,
) error {
	config, err := mod.FetchModuleConfiguration[ChannelConfig](ctx, db, moduleName)
	if err != nil {
		return err
	}

	m.SetConfig(config)
	m.logger.Infof(" -> channels: %+v", m.Config())
	return nil
}

func (m *Module) getChannelConfig(channelID string) (ChannelConfig, error) {
	for _, o := range m.Config() {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelConfig{}, errors.New("channel config not found")
}

func (m *Module) Start(ctx context.Context, discord *discordgo.Session, db *store.Queries) error {
	config, err := mod.FetchModuleConfiguration[ChannelConfig](
		ctx,
		db,
		moduleName,
	)

	if err != nil {
		return fmt.Errorf("unable to parse thisthat config: %w", err)
	}

	m.SetConfig(config)
	m.logger.Info("started thisthat module")
	m.logger.Infof(" -> channels: %+v", m.Config())
	m.registerHandlers(discord)
	return nil
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	discord.AddHandler(func(d *discordgo.Session, msg *discordgo.MessageCreate) {
		config, err := m.getChannelConfig(msg.ChannelID)

		if err != nil {
			return
		}

		if msg.ChannelID != config.ID || msg.Author.ID == discord.State.User.ID {
			return
		}

		num := getImageAttachmentCount(msg.Attachments)

		if num > 1 {
			for _, emoji := range emojis[:num] {
				d.MessageReactionAdd(msg.ChannelID, msg.ID, emoji)
			}
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
