package inatlookup

import (
	"context"
	"fmt"
	"log"
	"regexp"

	dg "github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"adamolsen.dev/buggins/internal/pkg/inatapi"
)

type commandHandler = func(*dg.Session, *dg.MessageCreate, string)

var inlineTaxaSearchRe = regexp.MustCompile(`(?m)\.(\w+ ?\w+?)\.`)

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

func InitFromEnv(d *dg.Session) (*Bot, error) {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, fmt.Errorf("inatlookup bot missing config: %v", err)
	}

	return New(d, c), nil
}

func (b *Bot) Start() {
	b.registerHandlers()
	log.Print("Started inatlookup bot")
}

func (b *Bot) registerHandlers() {
	if b.CommandPrefix == "" {
		b.CommandPrefix = ","
	}

	re, err := regexp.Compile(fmt.Sprintf(`(?m)^%s(\w+) +(.*)$`, b.CommandPrefix))

	if err != nil {
		log.Printf("error compiling command handler regex: %v", err)
		return
	}

	handlers := map[string]commandHandler{
		"t": b.lookupTaxa,
	}

	b.discord.AddHandler(func(d *dg.Session, m *dg.MessageCreate) {
		matches := re.FindStringSubmatch(m.Content)

		if matches != nil {
			command := matches[1]
			content := matches[2]

			handler, ok := handlers[command]

			if ok {
				handler(d, m, content)
			}
		}

		matches = inlineTaxaSearchRe.FindStringSubmatch(m.Content)

		if matches != nil {
			handler, ok := handlers["t"]

			if ok {
				handler(d, m, matches[1])
			}
		}
	})
}

func (b *Bot) lookupTaxa(d *dg.Session, m *dg.MessageCreate, content string) {
	r, err := b.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record
		p := message.NewPrinter(language.English)

		b.discord.ChannelMessageSendComplex(m.ChannelID, &dg.MessageSend{
			Embed: &dg.MessageEmbed{
				Thumbnail: &dg.MessageEmbedThumbnail{URL: r.DefaultPhoto.MediumUrl},
				Color:     5763719,
				Fields: []*dg.MessageEmbedField{
					{
						Value: fmt.Sprintf(
							"**[%s (%s)](https://inaturalist.org/taxa/%d)**",
							r.Name,
							r.PreferredCommonName,
							r.ID,
						),
						Inline: true,
					},
					{
						Name:  "Type",
						Value: cases.Title(language.English, cases.Compact).String(r.Rank),
					},
					{
						Name:  "Observers",
						Value: p.Sprintf("%d", r.ObservationCount),
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
