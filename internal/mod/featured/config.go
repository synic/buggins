package featured

import (
	"github.com/urfave/cli/v2"

	"github.com/synic/buggins/internal/mod"
)

var (
	guildID               string
	channelID             string
	requiredReactionCount int
)

type GuildConfig struct {
	ID                    string `json:"guild_id"`
	ChannelID             string `json:"channel_id"`
	RequiredReactionCount int    `json:"reaction_count"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "guild-id",
			Usage:       "Guild `GUILD_ID`",
			Aliases:     []string{"g"},
			Destination: &guildID,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "channel-id",
			Usage:       "Channel `CHANNEL_ID`",
			Aliases:     []string{"c"},
			Destination: &channelID,
			Required:    true,
		},
		&cli.IntFlag{
			Name:        "reaction-count",
			Aliases:     []string{"r"},
			Value:       6,
			Usage:       "Number of reactions to trigger `COUNT`",
			Destination: &requiredReactionCount,
		},
	}

	return mod.ConfigCommandOptions{
		Flags:      flags,
		KeyFlag:    "guild-id",
		ModuleName: moduleName,
		GetKey:     func() string { return guildID },
		GetData: func() any {
			return GuildConfig{
				ID:                    guildID,
				ChannelID:             channelID,
				RequiredReactionCount: requiredReactionCount,
			}
		},
	}
}
