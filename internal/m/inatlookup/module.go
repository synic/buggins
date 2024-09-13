package inatlookup

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/pkg/inat"
)

type commandHandler = func(*discordgo.Session, *discordgo.MessageCreate, string)

var inlineTaxaSearchRe = regexp.MustCompile(`(?m) \.(\w+ ?\w+?)\. `)

type GuildOptions struct {
	CommandPrefixRegex *regexp.Regexp
	Name               string   `mapstructure:"name"`
	ID                 string   `mapstructure:"id"`
	CommandPrefix      string   `mapstructure:"command_prefix"`
	Channels           []string `mapstructure:"channels"`
}

type Options struct {
	Guilds []GuildOptions `mapstructure:"guilds"`
}

type Module struct {
	api       inat.Api
	discord   *discordgo.Session
	options   Options
	isStarted bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *discordgo.Session, options Options) *Module {
	return &Module{options: options, api: inat.New(), discord: discord}
}

func Provider(
	c conf.Config,
	discord *discordgo.Session,
) (providerResult, error) {
	var options Options
	err := c.Populate("inatlookup", &options)

	if err != nil {
		return providerResult{}, err
	}

	return providerResult{Module: New(discord, options)}, nil
}

func (m *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range m.options.Guilds {
		if o.ID == guildID {
			return o, nil
		}
	}

	return GuildOptions{}, errors.New("guild not configured")
}

func (b *Module) Start() {
	if !b.isStarted {
		b.isStarted = true
		b.registerHandlers()
		log.Print("started inatlookup module")
	}
}

func (b *Module) registerHandlers() {
	for i, o := range b.options.Guilds {
		if o.CommandPrefix == "" {
			o.CommandPrefix = ","
		}
		re, err := regexp.Compile(fmt.Sprintf(`(?m)^%s(\w+) +(.*)$`, o.CommandPrefix))
		if err != nil {
			log.Printf("error compiling cmd prefix for guild %s: %v", o.ID, err)
			continue
		}
		o.CommandPrefixRegex = re
		if slices.Contains(o.Channels, "all") {
			o.Channels = []string{}
		}
		b.options.Guilds[i] = o
	}

	handlers := map[string]commandHandler{
		"t": b.lookupTaxa,
	}

	b.discord.AddHandler(func(d *discordgo.Session, m *discordgo.MessageCreate) {
		options, err := b.getGuildOptions(m.GuildID)

		if err != nil {
			return
		}

		if len(options.Channels) > 0 && !slices.Contains(options.Channels, m.ChannelID) {
			return
		}

		if options.CommandPrefixRegex == nil {
			log.Printf("guild %s does not have a valid command prefix", m.GuildID)
			return
		}

		matches := options.CommandPrefixRegex.FindStringSubmatch(m.Content)

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

func (b *Module) lookupTaxa(d *discordgo.Session, m *discordgo.MessageCreate, content string) {
	r, err := b.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record
		p := message.NewPrinter(language.English)

		b.discord.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: r.DefaultPhoto.MediumURL},
				Color:     5763719,
				Fields: []*discordgo.MessageEmbedField{
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
