package inat

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (b *bot) lookupTaxa(d *discordgo.Session, m *discordgo.MessageCreate, content string) {
	r, err := b.svc.searchTaxa(content)

	if err == nil && len(r.Results) > 0 {
		r := r.Results[0].Record

		b.Discord.ChannelMessageSendComplex(b.ChannelID, &discordgo.MessageSend{
			Content: fmt.Sprintf(
				"**[%s (%s)](https://inaturalist.org/taxa/%d)**",
				r.Name,
				r.PreferredCommonName,
				r.ID,
			),
			Embed: &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{URL: r.DefaultPhoto.MediumUrl},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Observer Count",
						Value: fmt.Sprintf("%d", r.ObservationCount),
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
