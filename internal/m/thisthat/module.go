package thisthat

import (
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
)

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

type ChannelOptions struct {
	ID string `mapstructure:"id"`
}

type Options struct {
	Channels []ChannelOptions `mapstructure:"channels"`
}

type Module struct {
	discord   *discordgo.Session
	options   Options
	isStarted bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *discordgo.Session, options Options) *Module {
	return &Module{options: options, discord: discord}
}

func Provider(c conf.Config,
	discord *discordgo.Session,
) (providerResult, error) {
	var options Options
	err := c.Populate("thisthat", &options)

	if err != nil {
		return providerResult{}, err
	}

	return providerResult{Module: New(discord, options)}, nil
}

func (m *Module) getChannelOptions(channelID string) (ChannelOptions, error) {
	for _, o := range m.options.Channels {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelOptions{}, errors.New("channel options not found")
}

func (b *Module) Start() {
	if !b.isStarted {
		b.isStarted = true
		log.Println("started thisthat module")
		b.registerHandlers()
	}
}

func (b *Module) registerHandlers() {
	b.discord.AddHandler(func(d *discordgo.Session, m *discordgo.MessageCreate) {
		options, err := b.getChannelOptions(m.ChannelID)

		if err != nil {
			return
		}

		if m.ChannelID != options.ID || m.Author.ID == b.discord.State.User.ID {
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
