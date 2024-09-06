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

	"adamolsen.dev/buggins/internal/pkg/inatapi"
	"adamolsen.dev/buggins/internal/store"
)

type BotConfig struct {
	CronPattern string `env:"INATOBS_CRON_PATTERN, default=0 * * * *"`
	ChannelID   string `env:"INATOBS_CHANNEL_ID, required"`
	ProjectID   string `env:"INATOBS_PROJECT_ID, required"`
	PageSize    int    `env:"INATOBS_PAGE_SIZE, default=10"`
}

type Bot struct {
	api                inatapi.Api
	discord            *dg.Session
	store              *store.Queries
	displayedObservers []int64
	BotConfig
	isStarted               bool
	slashCommandsRegistered bool
}

func New(discord *dg.Session, db *store.Queries, config BotConfig) *Bot {
	return &Bot{BotConfig: config, discord: discord, api: inatapi.New(), store: db}
}

func InitFromEnv(d *dg.Session, s *store.Queries) (*Bot, error) {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, fmt.Errorf("inatobs bot missing config: %w", err)
	}

	return New(d, s, c), nil
}

func (b *Bot) Start() {
	if !b.isStarted {
		log.Printf("Started inatobs bot with cron pattern '%s'", b.CronPattern)
		b.isStarted = true
		b.registerHandlers()
		c := cron.New()
		c.AddFunc(b.CronPattern, b.Post)
		c.Start()
	}
}

func (b *Bot) registerHandlers() {
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

func (b *Bot) registerSlashCommands() {
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

func (b *Bot) findUnseenObservation() (inatapi.Observation, error) {
	observations, err := b.api.FetchRecentProjectObservations(b.ProjectID, b.PageSize, 200)

	if len(observations) <= 0 {
		if err != nil {
			return inatapi.Observation{}, fmt.Errorf("error fetching observations: %w", err)
		}

		return inatapi.Observation{}, errors.New("no unseen observations found")
	}

	o, err := b.selectUnseenObservation(observations)

	if err != nil {
		return inatapi.Observation{}, fmt.Errorf("error fetching unseen observation: %w", err)
	}

	return o, nil
}

func (b *Bot) Post() {
	log.Print("Attempting to fetch an unseen observation to display")
	o, err := b.findUnseenObservation()

	if err != nil {
		log.Printf("error fetching unseen observation: %v", err)
		return
	}

	taxonName := "unknown"
	commonName := "unknown"
	taxon := o.Taxon

	if taxon.Name != "" {
		taxonName = taxon.Name

		if taxon.CommonName.Name != "" {
			commonName = taxon.CommonName.Name
		} else if taxon.DefaultName.Name != "" {
			commonName = taxon.DefaultName.Name
		} else if o.Species != "" {
			commonName = o.Species
		}
	}

	photos := o.Photos

	if len(photos) > 10 {
		photos = photos[:10]
	}

	files := make([]*dg.File, 0, len(photos))

	for _, photo := range photos {
		r, err := http.Get(photo.MediumUrl)

		if err != nil {
			log.Printf("unable to retrieve data for photo `%s`: %v", photo.MediumUrl, err)
			continue
		}

		defer r.Body.Close()
		files = append(files, &dg.File{
			Name:        photo.MediumUrl,
			ContentType: "image/jpeg",
			Reader:      r.Body,
		})
	}

	b.discord.ChannelMessageSendComplex(b.ChannelID, &dg.MessageSend{
		Content: fmt.Sprintf(
			"**[%s](https://inaturalist.org/people/%d) has spotted something new!**\n\n",
			o.Username,
			o.UserID,
		),
		Files: files,
		Embed: &dg.MessageEmbed{
			Fields: []*dg.MessageEmbedField{
				{
					Name:  "Taxon",
					Value: fmt.Sprintf("%s (%s)", taxonName, commonName),
				},
				{
					Name:  "iNaturalist Link",
					Value: fmt.Sprintf("https://inaturalist.org/observations/%d", o.ID),
				},
				{
					Name: "Our community iNaturalist Project",
					Value: fmt.Sprintf(
						"https://inaturalist.org/projects/%s",
						b.ProjectID,
					),
				},
			},
		},
	})

	log.Printf("Displaying observation id %d from %s", o.ID, o.Username)

	b.markObservationAsSeen(context.Background(), o)
}

func (b *Bot) markObservationAsSeen(
	ctx context.Context,
	o inatapi.Observation,
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

func (b *Bot) selectUnseenObservation(
	observations []inatapi.Observation,
) (inatapi.Observation, error) {
	var (
		observationIds     []int64
		unseen             []inatapi.Observation
		seenIds            []int64
		observerMap        = make(map[int64][]inatapi.Observation)
		potentialObservers []int64
	)

	for _, o := range observations {
		observationIds = append(observationIds, o.ID)
	}

	seen, err := b.store.FindObservationsByIds(context.Background(), observationIds)

	if err != nil {
		return inatapi.Observation{}, fmt.Errorf("error selecting seen observations: %w", err)
	}

	for _, o := range seen {
		seenIds = append(seenIds, o.ID)
	}

	for _, o := range observations {
		if !slices.Contains(seenIds, o.ID) {
			unseen = append(unseen, o)

			items, ok := observerMap[o.UserID]

			if !ok {
				items = make([]inatapi.Observation, 0)
			}

			items = append(items, o)
			observerMap[o.UserID] = items

			if !slices.Contains(b.displayedObservers, o.UserID) {
				potentialObservers = append(potentialObservers, o.UserID)
			}
		}
	}

	if len(unseen) <= 0 {
		return inatapi.Observation{}, errors.New("no unseen observations found")
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
		return inatapi.Observation{}, fmt.Errorf(
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