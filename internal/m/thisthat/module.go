package thisthat

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/synic/buggins/internal/store"
)

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

type Module struct {
	options     Options
	isStarted   bool
	optionsLock sync.RWMutex
}

func New(db *store.Queries) (*Module, error) {
	options, err := fetchModuleOptions(db)

	if err != nil {
		return nil, fmt.Errorf("unable to parse thisthat options: %w", err)
	}

	return &Module{options: options}, nil
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
		return err
	}

	m.SetOptions(options)
	log.Printf(" -> channels: %+v", m.Options().Channels)
	return nil
}

func (m *Module) getChannelOptions(channelID string) (ChannelOptions, error) {
	for _, o := range m.Options().Channels {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelOptions{}, errors.New("channel options not found")
}

func (m *Module) Start(discord *discordgo.Session) error {
	if !m.isStarted {
		m.isStarted = true
		log.Println("started thisthat module")
		log.Printf(" -> channels: %+v", m.Options().Channels)
		m.registerHandlers(discord)
	}
	return nil
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	discord.AddHandler(func(d *discordgo.Session, msg *discordgo.MessageCreate) {
		options, err := m.getChannelOptions(msg.ChannelID)

		if err != nil {
			return
		}

		if msg.ChannelID != options.ID || msg.Author.ID == discord.State.User.ID {
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
