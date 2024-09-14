package inatlookup

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/synic/buggins/internal/pkg/inat"
	"github.com/synic/buggins/internal/store"
)

type commandHandler = func(*discordgo.Session, *discordgo.MessageCreate, string)

var inlineTaxaSearchRe = regexp.MustCompile(`(?m) \.(\w+ ?\w+?)\. `)

type Module struct {
	api       inat.Api
	discord   *discordgo.Session
	options   Options
	isStarted bool
	mu        sync.Mutex
}

func New(discord *discordgo.Session, db *store.Queries) (*Module, error) {
	options, err := getModuleOptions(db)
	if err != nil {
		return &Module{}, err
	}

	return &Module{options: options, api: inat.New(), discord: discord}, nil
}

func (m *Module) getGuildOptions(guildID string) (GuildOptions, error) {
	for _, o := range m.GetOptions().Guilds {
		if o.ID == guildID {
			return o, nil
		}
	}

	return GuildOptions{}, errors.New("guild not configured")
}

func (m *Module) GetOptions() Options {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.options
}

func (m *Module) Start() {
	if !m.isStarted {
		m.isStarted = true
		m.registerHandlers()
		log.Print("started inatlookup module")
		log.Printf(" -> guilds: %+v", m.GetOptions().Guilds)
	}
}

func (m *Module) GetName() string {
	return moduleName
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

func (m *Module) registerHandlers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	handlers := map[string]commandHandler{
		"t": m.lookupTaxa,
	}

	m.discord.AddHandler(func(d *discordgo.Session, msg *discordgo.MessageCreate) {
		options, err := m.getGuildOptions(msg.GuildID)

		if err != nil {
			return
		}

		if len(options.Channels) > 0 && !slices.Contains(options.Channels, msg.ChannelID) {
			return
		}

		if options.CommandPrefixRegex == nil {
			log.Printf("guild %s does not have a valid command prefix", msg.GuildID)
			return
		}

		matches := options.CommandPrefixRegex.FindStringSubmatch(msg.Content)

		if matches != nil {
			command := matches[1]
			content := matches[2]

			handler, ok := handlers[command]

			if ok {
				handler(d, msg, content)
			}
		}

		matches = inlineTaxaSearchRe.FindStringSubmatch(msg.Content)

		if matches != nil {
			handler, ok := handlers["t"]

			if ok {
				handler(d, msg, matches[1])
			}
		}
	})
}

func (m *Module) lookupTaxa(d *discordgo.Session, msg *discordgo.MessageCreate, content string) {
	r, err := m.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record
		p := message.NewPrinter(language.English)

		m.discord.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
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
