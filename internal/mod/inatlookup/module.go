package inatlookup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/synic/buggins/internal/inat"
	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName         = "inatlookup"
	inlineTaxaSearchRe = regexp.MustCompile(`(?m) \.(\w+ ?\w+?)\. `)
)

type commandHandler = func(*discordgo.Session, *discordgo.MessageCreate, string)

type Module struct {
	api        inat.Api
	logger     *slog.Logger
	config     []GuildConfig
	configLock sync.RWMutex
}

func New(logger *slog.Logger) (*Module, error) {
	return &Module{api: inat.New(), logger: logger}, nil
}

func Provider(logger *slog.Logger) (mod.ModuleProviderResult, error) {
	module, err := New(logger.With("mod", moduleName))

	if err != nil {
		return mod.ModuleProviderResult{}, err
	}

	return mod.ModuleProviderResult{Module: module}, nil
}

func (m *Module) guildConfig(guildID string) (GuildConfig, error) {
	for _, c := range m.Config() {
		if c.ID == guildID {
			return c, nil
		}
	}

	return GuildConfig{}, errors.New("guild not configured")
}

func (m *Module) Config() []GuildConfig {
	m.configLock.RLock()
	defer m.configLock.RUnlock()
	return m.config
}

func (m *Module) SetConfig(config []GuildConfig) {
	m.configLock.Lock()
	defer m.configLock.Unlock()
	m.config = config

	for i, guild := range m.config {
		if guild.CommandPrefix == "" {
			guild.CommandPrefix = ","
		}

		re, err := regexp.Compile(fmt.Sprintf(`(?m)^%s(\w+) +(.*)$`, guild.CommandPrefix))

		if err != nil {
			m.logger.Info("error compiling command handler regex", "err", err)
			return
		}

		guild.CommandPrefixRegex = re
		m.config[i] = guild
	}
}

func (m *Module) Start(ctx context.Context, discord *discordgo.Session, db *store.Queries) error {
	config, err := mod.FetchModuleConfiguration[GuildConfig](ctx, db, moduleName)
	if err != nil {
		return err
	}
	m.SetConfig(config)
	m.registerHandlers(discord)
	m.logger.Info("started inatlookup module")
	m.logger.Info(" -> config", "guilds", m.Config())
	return nil
}

func (m *Module) Name() string {
	return moduleName
}

func (m *Module) ReloadConfig(
	ctx context.Context,
	discord *discordgo.Session,
	db *store.Queries,
) error {
	config, err := mod.FetchModuleConfiguration[GuildConfig](ctx, db, moduleName)
	if err != nil {
		return err
	}

	m.SetConfig(config)
	m.logger.Info(" -> guilds", "guilds", m.Config())
	return nil
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	m.configLock.Lock()
	defer m.configLock.Unlock()

	handlers := map[string]commandHandler{
		"t": m.lookupTaxa,
	}

	discord.AddHandler(func(d *discordgo.Session, msg *discordgo.MessageCreate) {
		config, err := m.guildConfig(msg.GuildID)

		if err != nil {
			return
		}

		if len(config.Channels) > 0 && !slices.Contains(config.Channels, msg.ChannelID) {
			return
		}

		if config.CommandPrefixRegex == nil {
			m.logger.Warn("guilddoes not have a valid command prefix", "guild", msg.GuildID)
			return
		}

		matches := config.CommandPrefixRegex.FindStringSubmatch(msg.Content)

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

func (m *Module) lookupTaxa(
	discord *discordgo.Session,
	msg *discordgo.MessageCreate,
	content string,
) {
	r, err := m.api.Search([]string{"taxa"}, content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record
		p := message.NewPrinter(language.English)

		discord.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
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
	} else {
		discord.ChannelMessageSend(msg.ChannelID, "Sorry, nothing could be found for that request")
	}
}
