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
	dg "github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/pkg/inat"
	"github.com/synic/buggins/internal/store"
)

type Config struct {
	CronPattern string `env:"INATOBS_CRON_PATTERN, default=0 * * * *"`
	ChannelID   string `env:"INATOBS_CHANNEL_ID, required"`
	ProjectID   string `env:"INATOBS_PROJECT_ID, required"`
	PageSize    int    `env:"INATOBS_PAGE_SIZE, default=10"`
}

type Module struct {
	Config
	api                     inat.Api
	discord                 *dg.Session
	store                   *store.Queries
	displayedObservers      []int64
	isStarted               bool
	slashCommandsRegistered bool
}

type providerResult struct {
	fx.Out
	Module m.Module `group:"modules"`
}

func New(discord *dg.Session, db *store.Queries, config Config) *Module {
	return &Module{Config: config, discord: discord, api: inat.New(), store: db}
}

func ProviderFromEnv(d *dg.Session, s *store.Queries) (providerResult, error) {
	var c Config

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return providerResult{}, fmt.Errorf("inatobs module missing config: %w", err)
	}

	return providerResult{Module: New(d, s, c)}, nil
}

func (b *Module) Start() {
	if !b.isStarted {
		log.Printf("Started inatobs module with cron pattern '%s'", b.CronPattern)
		b.isStarted = true
		b.registerHandlers()
		c := cron.New()
		c.AddFunc(b.CronPattern, b.Post)
		c.Start()
	}
}

func (b *Module) registerHandlers() {
	if b.discord.DataReady {
		b.registerSlashCommands()
	} else {
		b.discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
			log.Println("Discord connection detected, registering slash commands for inatobs")
			b.registerSlashCommands()
		})
	}

	b.discord.AddHandler(func(d *dg.Session, i *dg.InteractionCreate) {
		if i.ChannelID != b.ChannelID {
			d.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Wrong channel, bub.",
				},
			})
		}

		if i.ApplicationCommandData().Name == "loadinat" {
			log.Println("/loadinat called, loading observation to display")
			go b.Post()

			d.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Done, observation is loading and will be posted soon!",
				},
			})
		}
	})
}

func (b *Module) registerSlashCommands() {
	if !b.discord.DataReady {
		fmt.Println("Cannot register inatobs slash commands, websocket not yet connected")
		return
	}

	if b.slashCommandsRegistered {
		return
	}

	b.slashCommandsRegistered = true
	var adminPermissions int64 = discordgo.PermissionManageServer
	command := discordgo.ApplicationCommand{
		Name:                     "loadinat",
		Description:              "Load and display a random observation",
		DefaultMemberPermissions: &adminPermissions,
	}

	_, err := b.discord.ApplicationCommandCreate(b.discord.State.Application.ID, "", &command)

	if err != nil {
		log.Printf("error creating /loadinat command: %v", err)
	}

	log.Println("inatobs slash commands registered")
}

func (b *Module) findUnseenObservation() (inat.Observation, error) {
	observations, err := b.api.FetchRecentProjectObservations(b.ProjectID, b.PageSize, 200)

	if len(observations) <= 0 {
		if err != nil {
			return inat.Observation{}, fmt.Errorf("error fetching observations: %w", err)
		}

		return inat.Observation{}, errors.New("no unseen observations found")
	}

	o, err := b.selectUnseenObservation(observations)

	if err != nil {
		return inat.Observation{}, fmt.Errorf("error fetching unseen observation: %w", err)
	}

	return o, nil
}

func (b *Module) Post() {
	log.Print("Attempting to fetch an unseen observation to display")
	o, err := b.findUnseenObservation()

	if err != nil {
		log.Printf("error fetching unseen observation: %v", err)
		return
	}

	taxonName, commonName := o.GetTaxonNames()

	fields := make([]*dg.MessageEmbedField, 0)
	fields = append(fields, &dg.MessageEmbedField{
		Name:  "Taxon",
		Value: fmt.Sprintf("%s (%s)", taxonName, commonName),
	})
	fields = append(fields, &dg.MessageEmbedField{
		Name: "Our community iNaturalist Project",
		Value: fmt.Sprintf(
			"https://inaturalist.org/projects/%s",
			b.ProjectID,
		),
	})

	photos := o.Photos

	if len(photos) > 5 {
		photos = photos[:5]
	}

	files := make([]*dg.File, 0, len(photos))

	for _, photo := range photos {
		r, err := http.Get(photo.MediumURL)

		if err != nil {
			log.Printf("unable to retrieve data for photo `%s`: %v", photo.MediumURL, err)
			continue
		}

		defer r.Body.Close()
		files = append(files, &dg.File{
			Name:        photo.MediumURL,
			ContentType: "image/jpeg",
			Reader:      r.Body,
		})
	}

	b.discord.ChannelMessageSendComplex(b.ChannelID, &dg.MessageSend{
		Files: files,
		Embed: &dg.MessageEmbed{
			URL:   fmt.Sprintf("https://inaturalist.org/observations/%d", o.ID),
			Title: fmt.Sprintf("%s has spotted something new!", o.Username),
			Author: &dg.MessageEmbedAuthor{
				Name:    o.User.Username,
				URL:     fmt.Sprintf("https://inaturalist.org/people/%d", o.UserID),
				IconURL: o.User.UserIconURL,
			},
			Color:  2123412,
			Fields: fields,
		},
	})

	log.Printf("Displaying observation id %d from %s", o.ID, o.Username)

	b.markObservationAsSeen(context.Background(), o)
}

func (b *Module) markObservationAsSeen(
	ctx context.Context,
	o inat.Observation,
) (store.SeenObservation, error) {
	if !slices.Contains(b.displayedObservers, o.UserID) {
		b.displayedObservers = append(b.displayedObservers, o.UserID)
	}

	seen, err := b.store.CreateSeenObservation(ctx, o.ID)

	if err != nil {
		return store.SeenObservation{}, fmt.Errorf("error saving seen observation: %w", err)
	}

	return seen, nil
}

func (b *Module) selectUnseenObservation(
	observations []inat.Observation,
) (inat.Observation, error) {
	var (
		observationIds     []int64
		unseen             []inat.Observation
		seenIds            []int64
		observerMap        = make(map[int64][]inat.Observation)
		potentialObservers []int64
	)

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := b.store.FindObservationsByIds(context.Background(), observationIds)

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

			if !slices.Contains(b.displayedObservers, o.UserID) {
				potentialObservers = append(potentialObservers, o.UserID)
			}
		}
	}

	if len(unseen) <= 0 {
		return inat.Observation{}, errors.New("no unseen observations found")
	}

	if len(potentialObservers) <= 0 {
		potentialObservers = slices.Collect(maps.Keys(observerMap))
		b.displayedObservers = b.displayedObservers[:0]
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
