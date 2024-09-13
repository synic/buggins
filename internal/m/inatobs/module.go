package inatobs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/conf"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/pkg/inat"
	"github.com/synic/buggins/internal/store"
)

type ChannelOptions struct {
	ID          string `mapstructure:"id"`
	CronPattern string `mapstructure:"cron_pattern"`
	ProjectID   int64  `mapstructure:"inat_project_id"`
}

type Options struct {
	Channels []ChannelOptions `mapstructure:"channels"`
	PageSize int              `mapstructure:"page_size"`
}

type Module struct {
	api                     inat.Api
	discord                 *discordgo.Session
	store                   *store.Queries
	displayedObservers      map[string][]int64
	options                 Options
	isStarted               bool
	slashCommandsRegistered bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *discordgo.Session, db *store.Queries, options Options) *Module {
	return &Module{
		options:            options,
		discord:            discord,
		api:                inat.New(),
		store:              db,
		displayedObservers: make(map[string][]int64),
	}
}

func Provider(c conf.Config,
	discord *discordgo.Session,
	db *store.Queries,
) (providerResult, error) {
	var options Options
	err := c.Populate("inatobs", &options)

	if err != nil {
		return providerResult{}, err
	}

	return providerResult{Module: New(discord, db, options)}, nil
}

func (m *Module) Start() {
	if !m.isStarted {
		log.Println("started inatobs module")
		m.isStarted = true
		m.registerHandlers()

		for _, o := range m.options.Channels {
			pattern := o.CronPattern
			if pattern == "" {
				pattern = "0 * * * *"
			}

			c := cron.New()
			c.AddFunc(pattern, func() { m.Post(o.ID) })
			c.Start()
		}
	}
}

func (m *Module) getChannelOptions(channelID string) (ChannelOptions, error) {
	for _, o := range m.options.Channels {
		if o.ID == channelID {
			return o, nil
		}
	}

	return ChannelOptions{}, errors.New("channel options not found")
}

func (m *Module) registerHandlers() {
	if m.discord.DataReady {
		m.registerSlashCommands()
	} else {
		m.discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
			log.Println("Discord connection detected, registering slash commands for inatobs")
			m.registerSlashCommands()
		})
	}

	m.discord.AddHandler(func(d *discordgo.Session, i *discordgo.InteractionCreate) {
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
			log.Println("/loadinat called, loading observation to display")
			go m.Post(i.ChannelID)

			d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Done, observation is loading and will be posted soon!",
				},
			})
		}
	})
}

func (m *Module) registerSlashCommands() {
	if !m.discord.DataReady {
		fmt.Println("Cannot register inatobs slash commands, websocket not yet connected")
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

	_, err := m.discord.ApplicationCommandCreate(m.discord.State.Application.ID, "", &command)

	if err != nil {
		log.Printf("error creating /loadinat command: %v", err)
	}

	log.Println("inatobs slash commands registered")
}

func (m *Module) findUnseenObservation(
	channelID string,
	projectID int64,
) (inat.Observation, error) {
	options, err := m.getChannelOptions(channelID)

	if err != nil {
		return inat.Observation{}, err
	}

	observations, err := m.api.FetchRecentProjectObservations(
		options.ProjectID,
		m.options.PageSize,
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

func (m *Module) Post(channelID string) {
	options, err := m.getChannelOptions(channelID)

	if err != nil {
		return
	}

	log.Print("Attempting to fetch an unseen observation to display")
	o, err := m.findUnseenObservation(channelID, options.ProjectID)

	if err != nil {
		log.Printf("error fetching unseen observation: %v", err)
		return
	}

	taxonName, commonName := o.GetTaxonNames()

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
		r, err := http.Get(photo.MediumURL)

		if err != nil {
			log.Printf("unable to retrieve data for photo `%s`: %v", photo.MediumURL, err)
			continue
		}

		defer r.Body.Close()
		files = append(files, &discordgo.File{
			Name:        photo.MediumURL,
			ContentType: "image/jpeg",
			Reader:      r.Body,
		})
	}

	m.discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
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

	log.Printf("Displaying observation id %d from %s", o.ID, o.Username)

	m.markObservationAsSeen(context.Background(), channelID, o)
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

	displayed, ok := m.displayedObservers[channelID]

	if !ok {
		displayed = make([]int64, 0)
	}

	if !slices.Contains(displayed, o.UserID) {
		displayed = append(displayed, o.UserID)
	}

	m.displayedObservers[channelID] = displayed
	seen, err := m.store.CreateSeenObservation(
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

	displayed, ok := m.displayedObservers[channelID]

	if !ok {
		displayed = make([]int64, 0)
	}

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := m.store.FindObservations(context.Background(), store.FindObservationsParams{
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
		m.displayedObservers[channelID] = displayed
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
