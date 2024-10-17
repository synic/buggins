package inatlookup

import (
	"regexp"
	"slices"

	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/mod"
)

var (
	guildID       string
	commandPrefix string
	channels      []string
)

type GuildConfig struct {
	CommandPrefixRegex *regexp.Regexp `json:"-"`
	Name               string         `json:"name"`
	ID                 string         `json:"id"`
	CommandPrefix      string         `json:"command_prefix"`
	Channels           []string       `json:"channels"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&guildID, "guild-id", "g", "", "Guild ID")
	flags.StringVar(&commandPrefix, "command-prefix", ",", "Command prefix")
	flags.StringArrayVarP(&channels, "channels", "c", []string{}, "Channels")

	return mod.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "guild-id",
		RequiredFlags: []string{"guild-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return guildID },
		GetData: func() any {
			c := channels[:]

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
