package thisthat

import (
	"context"
	"log"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
)

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

type BotConfig struct {
	ChannelID string `env:"THISTHAT_CHANNEL_ID, required"`
}

type Bot struct {
	BotConfig
	discord            *dg.Session
	handlersRegistered bool
}

func getImageAttachmentCount(attachments []*dg.MessageAttachment) int {
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

func New(discord *dg.Session, config BotConfig) *Bot {
	return &Bot{BotConfig: config, discord: discord}
}

func (b *Bot) Start() {
	log.Println("Started thisthat bot...")
	b.registerHandlers()
}

func (b *Bot) registerHandlers() {
	b.discord.AddHandler(func(d *dg.Session, m *dg.MessageCreate) {
		if m.ChannelID != b.ChannelID {
			return
		}

		num := getImageAttachmentCount(m.Attachments)

		if num > 1 {
			for _, emoji := range emojis[:num] {
				d.MessageReactionAdd(m.ChannelID, m.ID, emoji)
			}
		}
	})
}

func InitFromEnvironment(d *dg.Session) *Bot {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Printf("thisthat bot missing config, disabled.: %v\n", err)
		return nil
	}

	return New(d, c)
}
