package inatlookup

import (
	"context"
	"fmt"
	"log"
	"regexp"

	dg "github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/pkg/inat"
)

type commandHandler = func(*dg.Session, *dg.MessageCreate, string)

var inlineTaxaSearchRe = regexp.MustCompile(`(?m) \.(\w+ ?\w+?)\. `)

type Config struct {
	CommandPrefix string `env:"INATLOOKUP_COMMAND_PREFIX, default=,"`
}

type Module struct {
	Config
	api       inat.Api
	discord   *dg.Session
	isStarted bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *dg.Session, config Config) *Module {
	return &Module{Config: config, api: inat.New(), discord: discord}
}

func ProviderFromEnv(d *dg.Session) (providerResult, error) {
	var c Config

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return providerResult{}, fmt.Errorf("inatlookup module missing config: %w", err)
	}

	return providerResult{Module: New(d, c)}, nil
}

func (b *Module) Start() {
	if !b.isStarted {
		b.isStarted = true
		b.registerHandlers()
		log.Print("Started inatlookup module")
	}
}

func (b *Module) registerHandlers() {
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

func (b *Module) lookupTaxa(d *dg.Session, m *dg.MessageCreate, content string) {
	r, err := b.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record
		p := message.NewPrinter(language.English)

		b.discord.ChannelMessageSendComplex(m.ChannelID, &dg.MessageSend{
			Embed: &dg.MessageEmbed{
				Thumbnail: &dg.MessageEmbedThumbnail{URL: r.DefaultPhoto.MediumURL},
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
