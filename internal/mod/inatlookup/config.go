package inatlookup

import (
	"regexp"
	"slices"

	"github.com/urfave/cli/v2"

	"github.com/synic/buggins/internal/mod"
)

var (
	guildID       string
	commandPrefix string
	channels      cli.StringSlice
)

type GuildConfig struct {
	CommandPrefixRegex *regexp.Regexp `json:"-"`
	Name               string         `json:"name"`
	ID                 string         `json:"id"`
	CommandPrefix      string         `json:"command_prefix"`
	Channels           []string       `json:"channels"`
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
			Name:        "command-prefix",
			Usage:       "Command prefix `PREFIX`",
			Destination: &commandPrefix,
			Value:       ",",
		},
		&cli.StringSliceFlag{
			Name:        "channels",
			Usage:       "Channel ids (omit for all channels) `CHANNEL_IDS`",
			Aliases:     []string{"c"},
			Destination: &channels,
			Value:       cli.NewStringSlice(),
		},
	}

	return mod.ConfigCommandOptions{
		Flags:      flags,
		KeyFlag:    "guild-id",
		ModuleName: moduleName,
		GetKey:     func() string { return guildID },
		GetData: func() any {
			c := channels.Value()[:]

			if slices.Contains(c, "all") {
				c = []string{}
			}

			return GuildConfig{
				ID:            guildID,
				CommandPrefix: commandPrefix,
				Channels:      c,
			}
		},
	}
}
