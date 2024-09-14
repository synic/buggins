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
	discord   *discordgo.Session
	options   Options
	isStarted bool
	mu        sync.Mutex
}

func New(discord *discordgo.Session, db *store.Queries) (*Module, error) {
	options, err := getModuleOptions(db)

	if err != nil {
		return nil, fmt.Errorf("unable to parse thisthat options: %w", err)
	}

	return &Module{options: options, discord: discord}, nil
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
		return err
	}

	m.options = options
	return nil
}

func (m *Module) getChannelOptions(channelID string) (ChannelOptions, error) {
	for _, o := range m.GetOptions().Channels {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelOptions{}, errors.New("channel options not found")
}

func (m *Module) Start() {
	if !m.isStarted {
		m.isStarted = true
		log.Println("started thisthat module")
		log.Printf(" -> channels: %+v", m.GetOptions().Channels)
		m.registerHandlers()
	}
}

func (m *Module) registerHandlers() {
	m.discord.AddHandler(func(d *discordgo.Session, msg *discordgo.MessageCreate) {
		options, err := m.getChannelOptions(msg.ChannelID)

		if err != nil {
			return
		}

		if msg.ChannelID != options.ID || msg.Author.ID == m.discord.State.User.ID {
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
