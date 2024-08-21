package thisthat

import (
	"log"
	"strings"

	dg "github.com/bwmarrin/discordgo"
)

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

type BotConfig struct {
	Discord   *dg.Session
	ChannelID string
}

type bot struct {
	BotConfig
	isRunning bool
}

func shouldReact(attachments []*dg.MessageAttachment) bool {
	if len(attachments) <= 1 {
		return false
	}

	shouldReact := true

	for _, attachment := range attachments {
		if !strings.Contains(attachment.ContentType, "image") {
			shouldReact = false
		}
	}

	return shouldReact
}

func New(config BotConfig) *bot {
	b := &bot{BotConfig: config}
	b.registerHandlers()
	return b
}

func (b *bot) Start() {
	log.Println("Starting thisthat bot...")
	b.isRunning = true
}

func (b *bot) Stop() {
	b.isRunning = false
	log.Println("Stopping thisthat bot...")
}

func (b *bot) registerHandlers() {
	b.Discord.AddHandler(func(d *dg.Session, m *dg.MessageCreate) {
		if !b.isRunning || m.ChannelID != b.ChannelID {
			return
		}

		if shouldReact(m.Attachments) {
			num := len(m.Attachments)

			for _, emoji := range emojis[:num] {
				d.MessageReactionAdd(m.ChannelID, m.ID, emoji)
			}
		}
	})
}
