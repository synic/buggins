package thisthat

import (
	"github.com/urfave/cli/v2"

	"github.com/synic/buggins/internal/mod"
)

var (
	channelID string
)

type ChannelConfig struct {
	ID string `json:"id"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "channel-id",
			Aliases:     []string{"c"},
			Required:    true,
			Usage:       "Channel `CHANNEL_ID`",
			Destination: &channelID,
		},
	}

	return mod.ConfigCommandOptions{
		ModuleName: moduleName,
		KeyFlag:    "channel-id",
		Flags:      flags,
		GetKey:     func() string { return channelID },
		GetData: func() any {
			return ChannelConfig{
				ID: channelID,
			}
		},
	}
}
