package inatobs

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"net/http"
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/robfig/cron/v3"

	"github.com/synic/buggins/internal/inat"
	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "inatobs"
)

type Module struct {
	api                     inat.Api
	logger                  *log.Logger
	db                      *store.Queries
	displayedObservers      map[string][]int64
	config                  []ChannelConfig
	crons                   []*cron.Cron
	slashCommandsRegistered bool
	configLock              sync.RWMutex
	cronsLock               sync.Mutex
	displayedObserversLock  sync.RWMutex
}

func New(db *store.Queries, logger *log.Logger) (*Module, error) {
	return &Module{
		api:                inat.New(),
		db:                 db,
		logger:             logger,
		displayedObservers: make(map[string][]int64),
		crons:              make([]*cron.Cron, 0),
	}, nil
}

func Provider(db *store.Queries, logger *log.Logger) (mod.ModuleProviderResult, error) {
	module, err := New(db, logger.With("mod", moduleName))

	if err != nil {
		return mod.ModuleProviderResult{}, err
	}

	return mod.ModuleProviderResult{Module: module}, nil
}

func (m *Module) Config() []ChannelConfig {
	m.configLock.RLock()
	defer m.configLock.RUnlock()
	return m.config
}

func (m *Module) SetConfig(config []ChannelConfig) {
	m.configLock.Lock()
	defer m.configLock.Unlock()
	m.config = config
}

func (m *Module) Start(ctx context.Context, discord *discordgo.Session, db *store.Queries) error {
	config, err := mod.FetchModuleConfiguration[ChannelConfig](ctx, db, moduleName)
	if err != nil {
		return err
	}
	m.SetConfig(config)
	m.logger.Info("started inatobs module")
	m.logger.Infof(" -> channels: %+v", m.Config())
	m.registerHandlers(discord)
	m.startCrons(discord)
	return nil
}

func (m *Module) startCrons(discord *discordgo.Session) {
	m.cronsLock.Lock()
	defer m.cronsLock.Unlock()
	for _, c := range m.crons {
		c.Stop()
	}
	m.crons = make([]*cron.Cron, 0, len(m.Config()))
	for _, o := range m.Config() {
		pattern := o.CronPattern
		c := cron.New()
		c.AddFunc(pattern, func() { m.Post(discord, o.ID) })
		c.Start()
		m.crons = append(m.crons, c)
	}
}

func (m *Module) Name() string {
	return moduleName
}

func (m *Module) ReloadConfig(
	ctx context.Context,
	discord *discordgo.Session,
	db *store.Queries,
) error {
	config, err := mod.FetchModuleConfiguration[ChannelConfig](ctx, db, moduleName)
	if err != nil {
		return err
	}

	m.SetConfig(config)
	m.startCrons(discord)
	m.logger.Infof(" -> channels: %+v", m.Config())

	return nil
}

func (m *Module) getChannelOptions(channelID string) (ChannelConfig, error) {
	for _, o := range m.Config() {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelConfig{}, errors.New("channel config not found")
}

func (m *Module) registerHandlers(discord *discordgo.Session) {
	if discord.DataReady {
		m.registerSlashCommands(discord)
	} else {
		discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
			m.logger.Info(" -> discord connection detected, registering slash commands for inatobs")
			m.registerSlashCommands(discord)
		})
	}

	discord.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
		_, err := m.getChannelOptions(i.ChannelID)

		if err != nil {
			d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Wrong channel, bub.",
				},
			})
		}

		if i.ApplicationCommandData().Name == "loadinat" {
			m.logger.Info("/loadinat called, loading observation to display")
			go m.Post(discord, i.ChannelID)

			d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Done, observation is loading and will be posted soon!",
				},
			})
		}
	})
}

func (m *Module) registerSlashCommands(discord *discordgo.Session) {
	if !discord.DataReady {
		fmt.Println("cannot register inatobs slash commands, websocket not yet connected")
		return
	}

	if m.slashCommandsRegistered {
		return
	}

	m.slashCommandsRegistered = true
	var adminPermissions int64 = discordgo.PermissionManageServer
	command := discordgo.ApplicationCommand{
		Name:                     "loadinat",
		Description:              "Load and display a random observation",
		DefaultMemberPermissions: &adminPermissions,
	}

	_, err := discord.ApplicationCommandCreate(discord.State.Application.ID, "", &command)

	if err != nil {
		m.logger.Warnf("error creating /loadinat command: %v", err)
	}

	m.logger.Info(" -> inatobs slash commands registered")
}

func (m *Module) findUnseenObservation(
	channelID string,
	projectID int64,
) (inat.Observation, error) {
	config, err := m.getChannelOptions(channelID)

	if err != nil {
		return inat.Observation{}, err
	}

	observations, err := m.api.FetchRecentProjectObservations(
		config.ProjectID,
		config.PageSize,
		200,
	)

	if len(observations) <= 0 {
		if err != nil {
			return inat.Observation{}, fmt.Errorf("error fetching observations: %w", err)
		}

		return inat.Observation{}, errors.New("no unseen observations found")
	}

	o, err := m.selectUnseenObservation(channelID, projectID, observations)

	if err != nil {
		return inat.Observation{}, fmt.Errorf("error fetching unseen observation: %w", err)
	}

	return o, nil
}

func (m *Module) Post(discord *discordgo.Session, channelID string) {
	options, err := m.getChannelOptions(channelID)

	if err != nil {
		return
	}

	m.logger.Info("Attempting to fetch an unseen observation to display")
	o, err := m.findUnseenObservation(channelID, options.ProjectID)

	if err != nil {
		m.logger.Errorf("error fetching unseen observation: %v", err)
		return
	}

	taxonName, commonName := o.TaxonNames()

	fields := make([]*discordgo.MessageEmbedField, 0)
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Taxon",
		Value: fmt.Sprintf("%s (%s)", taxonName, commonName),
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Our community iNaturalist Project",
		Value: fmt.Sprintf("https://inaturalist.org/projects/%d", options.ProjectID),
	})

	photos := o.Photos

	if len(photos) > 5 {
		photos = photos[:5]
	}

	files := make([]*discordgo.File, 0, len(photos))

	for _, photo := range photos {
		res, err := http.Get(photo.MediumURL)

		if err != nil {
			m.logger.Errorf("unable to retrieve data for photo `%s`: %v", photo.MediumURL, err)
			continue
		}

		defer res.Body.Close()
		files = append(files, &discordgo.File{
			Name:        photo.MediumURL,
			ContentType: "image/jpeg",
			Reader:      res.Body,
		})
	}

	discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Files: files,
		Embed: &discordgo.MessageEmbed{
			URL:   fmt.Sprintf("https://inaturalist.org/observations/%d", o.ID),
			Title: fmt.Sprintf("%s has spotted something new!", o.Username),
			Author: &discordgo.MessageEmbedAuthor{
				Name:    o.User.Username,
				URL:     fmt.Sprintf("https://inaturalist.org/people/%d", o.UserID),
				IconURL: o.User.UserIconURL,
			},
			Color:  2123412,
			Fields: fields,
		},
	})

	m.logger.Infof("Displaying observation id %d from %s", o.ID, o.Username)

	m.markObservationAsSeen(context.Background(), channelID, o)
}

func (m *Module) getDisplayedObservers(channelID string) ([]int64, bool) {
	m.displayedObserversLock.RLock()
	defer m.displayedObserversLock.RUnlock()
	items, ok := m.displayedObservers[channelID]
	return items, ok
}

func (m *Module) setDisplayedObservers(channelID string, do []int64) {
	m.displayedObserversLock.Lock()
	defer m.displayedObserversLock.Unlock()
	m.displayedObservers[channelID] = do
}

func (m *Module) markObservationAsSeen(
	ctx context.Context,
	channelID string,
	o inat.Observation,
) (store.SeenObservation, error) {
	options, err := m.getChannelOptions(channelID)

	if err != nil {
		return store.SeenObservation{}, err
	}

	displayed, ok := m.getDisplayedObservers(channelID)

	if !ok {
		displayed = make([]int64, 0)
	}

	if !slices.Contains(displayed, o.UserID) {
		displayed = append(displayed, o.UserID)
	}

	m.setDisplayedObservers(channelID, displayed)
	seen, err := m.db.CreateSeenObservation(
		ctx,
		store.CreateSeenObservationParams{
			ID:        o.ID,
			ProjectID: options.ProjectID,
			ChannelID: channelID,
		},
	)

	if err != nil {
		return store.SeenObservation{}, fmt.Errorf("error saving seen observation: %w", err)
	}

	return seen, nil
}

func (m *Module) selectUnseenObservation(
	channelID string,
	projectID int64,
	observations []inat.Observation,
) (inat.Observation, error) {
	var (
		observationIds     []int64
		unseen             []inat.Observation
		seenIds            []int64
		observerMap        = make(map[int64][]inat.Observation)
		potentialObservers []int64
	)

	displayed, ok := m.getDisplayedObservers(channelID)

	if !ok {
		displayed = make([]int64, 0)
	}

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := m.db.FindObservations(context.Background(), store.FindObservationsParams{
		ID:        observationIds,
		ProjectID: projectID,
		ChannelID: channelID,
	})

	if err != nil {
		return inat.Observation{}, fmt.Errorf("error selecting seen observations: %w", err)
	}

	for _, o := range seen {
		seenIds = append(seenIds, o.ID)
	}

	for _, o := range observations {
		if !slices.Contains(seenIds, o.ID) {
			unseen = append(unseen, o)

			items, ok := observerMap[o.UserID]

			if !ok {
				items = make([]inat.Observation, 0)
			}

			items = append(items, o)
			observerMap[o.UserID] = items

			if !slices.Contains(displayed, o.UserID) {
				potentialObservers = append(potentialObservers, o.UserID)
			}
		}
	}

	if len(unseen) <= 0 {
		return inat.Observation{}, errors.New("no unseen observations found")
	}

	if len(potentialObservers) <= 0 {
		potentialObservers = slices.Collect(maps.Keys(observerMap))
		displayed = displayed[:0]
		m.setDisplayedObservers(channelID, displayed)
	}

	rand.Shuffle(len(potentialObservers), func(i, j int) {
		potentialObservers[i], potentialObservers[j] = potentialObservers[j], potentialObservers[i]
	})

	observerId := potentialObservers[0]
	items, ok := observerMap[observerId]

	if !ok || len(items) <= 0 {
		return inat.Observation{}, fmt.Errorf(
			"could not find unseen observations for observer %d",
			observerId,
		)
	}

	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	observation := items[0]

	return observation, nil

}
