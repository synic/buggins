package inat

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	dg "github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"

	"adamolsen.dev/buggins/internal/store"
)

type BotConfig struct {
	CronPattern string
	Discord     *dg.Session
	ChannelID   string
	GuildID     string
	ProjectID   string
	Store       *store.Queries
	PageSize    int
}

type bot struct {
	BotConfig
	svc       service
	job       *cron.Cron
	isRunning bool
}

func New(config BotConfig) *bot {
	s := newService(
		serviceConfig{
			projectID: config.ProjectID,
			pageSize:  config.PageSize,
			store:     config.Store,
		},
	)
	b := bot{BotConfig: config, svc: s}
	b.registerHandlers()
	return &b
}

func (b *bot) Start() {
	b.Stop()
	b.job = cron.New()
	b.job.AddFunc(b.CronPattern, b.Post)
	b.job.Start()
	b.isRunning = true
	log.Printf("Started inat bot with cron pattern '%s'...", b.CronPattern)
}

func (b *bot) Stop() {
	b.isRunning = false
	if b.job != nil {
		b.job.Stop()
		b.job = nil
		log.Print("Stopped inat bot...")
	}
}

func (b *bot) registerHandlers() {
	b.Discord.AddHandler(func(d *dg.Session, r *dg.Ready) {
		b.registerSlashCommands()
	})

	b.Discord.AddHandler(func(d *dg.Session, i *dg.InteractionCreate) {
		if !b.isRunning {
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

func (b *bot) registerSlashCommands() {
	var adminPermissions int64 = discordgo.PermissionManageServer
	command := discordgo.ApplicationCommand{
		Name:                     "loadinat",
		Description:              "Load and display a random observation",
		DefaultMemberPermissions: &adminPermissions,
	}

	_, err := b.Discord.ApplicationCommandCreate(b.Discord.State.User.ID, b.GuildID, &command)

	if err != nil {
		log.Printf("error creating /loadinat command: %v", err)
	}
}

func (b *bot) Post() {
	if !b.isRunning {
		return
	}

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

	b.Discord.ChannelMessageSendComplex(b.ChannelID, &dg.MessageSend{
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
