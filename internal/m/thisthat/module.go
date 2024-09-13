package thisthat

import (
	"context"
	"fmt"
	"log"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/m"
)

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

type Config struct {
	ChannelID string `env:"THISTHAT_CHANNEL_ID, required"`
}

type Module struct {
	Config
	discord   *dg.Session
	isStarted bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *dg.Session, config Config) *Module {
	return &Module{Config: config, discord: discord}
}

func ProviderFromEnv(d *dg.Session) (providerResult, error) {
	var c Config

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return providerResult{}, fmt.Errorf("thisthat module missing config: %w", err)
	}

	return providerResult{Module: New(d, c)}, nil
}

func (b *Module) Start() {
	if !b.isStarted {
		b.isStarted = true
		log.Println("Started thisthat module")
		b.registerHandlers()
	}
}

func (b *Module) registerHandlers() {
	b.discord.AddHandler(func(d *dg.Session, m *dg.MessageCreate) {
		if m.ChannelID != b.ChannelID || m.Author.ID == b.discord.State.User.ID {
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
