package inatlookup

import (
	"context"
	"fmt"
	"log"
	"regexp"

	dg "github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"

	"adamolsen.dev/buggins/internal/pkg/inatapi"
)

type commandHandler = func(*dg.Session, *dg.MessageCreate, string)

type BotConfig struct {
	CommandPrefix string `env:"INATLOOKUP_COMMAND_PREFIX, default=,"`
}

type Bot struct {
	BotConfig
	discord *dg.Session
	api     inatapi.Api
}

func New(discord *dg.Session, config BotConfig) *Bot {
	return &Bot{BotConfig: config, api: inatapi.New(), discord: discord}
}

func (b *Bot) Start() {
	b.registerHandlers()
	log.Print("Started inatlookup bot")
}

func (b *Bot) registerHandlers() {
	if b.CommandPrefix == "" {
		b.CommandPrefix = ","
	}

	commandRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^%s(\w+) +(.*)$`, b.CommandPrefix))
	commandHandlers := map[string]commandHandler{
		"t": b.lookupTaxa,
	}

	b.discord.AddHandler(func(d *dg.Session, m *dg.MessageCreate) {
		matches := commandRegex.FindStringSubmatch(m.Content)

		if matches == nil {
			return
		}

		command := matches[1]
		rest := matches[2]

		handler, ok := commandHandlers[command]

		if ok {
			handler(d, m, rest)
		}
	})
}

func (b *Bot) lookupTaxa(d *dg.Session, m *dg.MessageCreate, content string) {
	r, err := b.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record

		b.discord.ChannelMessageSendComplex(m.ChannelID, &dg.MessageSend{
			Content: fmt.Sprintf(
				"**[%s (%s)](https://inaturalist.org/taxa/%d)**",
				r.Name,
				r.PreferredCommonName,
				r.ID,
			),
			Embed: &dg.MessageEmbed{
				Image: &dg.MessageEmbedImage{URL: r.DefaultPhoto.MediumUrl},
				Fields: []*dg.MessageEmbedField{
					{
						Name:  "Observer Count",
						Value: fmt.Sprintf("%d", r.ObservationCount),
					},
					{
						Name:  "iNaturalist Link",
						Value: fmt.Sprintf("https://inaturalist.org/taxa/%d", r.ID),
					},
				},
			},
		})
	}
}

func InitFromEnvironment(d *dg.Session) *Bot {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Printf("inatlookup bot missing config, disabled.: %v\n", err)
		return nil
	}

	return New(d, c)
}
