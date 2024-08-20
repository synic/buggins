package inatobs

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	dg "github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	"github.com/sethvargo/go-envconfig"

	"adamolsen.dev/buggins/internal/store"
)

type BotConfig struct {
	CronPattern string `env:"INATOBS_CRON_PATTERN, default=0 * * * *"`
	ChannelID   string `env:"INATOBS_CHANNEL_ID, required"`
	ProjectID   string `env:"INATOBS_PROJECT_ID, required"`
	PageSize    int    `env:"INATOBS_PAGE_SIZE, default=10"`
}

type Bot struct {
	BotConfig
	discord            *dg.Session
	svc                service
	handlersRegistered bool
}

func New(discord *dg.Session, db *store.Queries, config BotConfig) *Bot {
	s := newService(
		serviceConfig{
			projectID: config.ProjectID,
			pageSize:  config.PageSize,
			store:     db,
		},
	)
	return &Bot{BotConfig: config, svc: s, discord: discord}
}

func (b *Bot) Start() {
	b.registerHandlers()
	c := cron.New()
	c.AddFunc(b.CronPattern, b.Post)
	c.Start()
	log.Printf("Started inatobs bot with cron pattern '%s'...", b.CronPattern)
}

func (b *Bot) registerHandlers() {
	b.discord.AddHandler(func(d *dg.Session, r *dg.Ready) {
		b.registerSlashCommands()
	})

	b.discord.AddHandler(func(d *dg.Session, i *dg.InteractionCreate) {
		if i.ChannelID != b.ChannelID {
			return
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
	var adminPermissions int64 = discordgo.PermissionManageServer
	command := discordgo.ApplicationCommand{
		Name:                     "loadinat",
		Description:              "Load and display a random observation",
		DefaultMemberPermissions: &adminPermissions,
	}

	_, err := b.discord.ApplicationCommandCreate(b.discord.State.User.ID, "", &command)

	if err != nil {
		log.Printf("error creating /loadinat command: %v", err)
	}
}

func (b *Bot) Post() {
	log.Print("Attempting to fetch an unseen observation to display")
	o, err := b.svc.FindUnseenObservation()

	if err != nil {
		log.Printf("error fetching unseen observation: %v", err)
		return
	}

	taxonName := "unknown"
	commonName := "unknown"
	taxon := o.Taxon

	if taxon != nil && taxon.Name != nil {
		if taxon.Name != nil {
			taxonName = *taxon.Name
		}

		if taxon.CommonName != nil && taxon.CommonName.Name != nil {
			commonName = *taxon.CommonName.Name
		} else if taxon.DefaultName != nil && taxon.DefaultName.Name != nil {
			commonName = *taxon.DefaultName.Name
		} else if o.Species != nil {
			commonName = *o.Species
		}
	}

	b.discord.ChannelMessageSendComplex(b.ChannelID, &dg.MessageSend{
		Content: fmt.Sprintf(
			"**[%s](https://inaturalist.org/people/%d) has spotted something new!**",
			o.Username,
			o.UserID,
		),
		Embed: &dg.MessageEmbed{
			Image: &dg.MessageEmbedImage{URL: o.Photos[0].LargeUrl},
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
						b.svc.projectID,
					),
				},
			},
		},
	})

	log.Printf("Displaying observation id %d from %s", o.ID, o.Username)

	b.svc.MarkObservationAsSeen(context.Background(), o)
}

func InitFromEnvironment(d *dg.Session, s *store.Queries) *Bot {
	var c BotConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Printf("inatobs bot missing config, disabled.: %v\n", err)
		return nil
	}

	return New(d, s, c)
}
