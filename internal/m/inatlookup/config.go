package inatlookup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"slices"

	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "inatlookup"

	guildID       string
	commandPrefix string
	channels      []string
)

type GuildOptions struct {
	CommandPrefixRegex *regexp.Regexp `json:"-"`
	Name               string         `json:"name"`
	ID                 string         `json:"id"`
	CommandPrefix      string         `json:"command_prefix"`
	Channels           []string       `json:"channels"`
}

type Options struct {
	Guilds []GuildOptions `json:"guilds"`
}

func fetchModuleOptions(db *store.Queries) (Options, error) {
	ctx := context.Background()
	rows, err := db.FindModuleConfigurations(ctx, moduleName)

	if err != nil {
		return Options{}, err
	}

	guilds := make([]GuildOptions, 0, len(rows))

	for _, row := range rows {
		var options GuildOptions

		data, ok := row.Options.([]byte)

		if !ok {
			log.Printf("could not parse inatlookup configuration for guild %s", row.Key)
			continue
		}
		err := json.Unmarshal(data, &options)

		if err != nil {
			log.Printf("could not parse inatlookup configuration for guild %s: %v", row.Key, err)
			continue
		}

		re, err := regexp.Compile(fmt.Sprintf(`(?m)^%s(\w+) +(.*)$`, options.CommandPrefix))
		if err != nil {
			return Options{}, err
		}

		options.CommandPrefixRegex = re

		guilds = append(guilds, options)
	}

	return Options{Guilds: guilds}, nil
}

func GetConfigCommandOptions() m.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&guildID, "guild-id", "g", "", "Guild ID")
	flags.StringVar(&commandPrefix, "command-prefix", ",", "Command prefix")
	flags.StringArrayVarP(&channels, "channels", "c", []string{}, "Channels")

	return m.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "guild-id",
		RequiredFlags: []string{"guild-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return guildID },
		GetData: func() any {
			c := channels

			if slices.Contains(c, "all") {
				c = []string{}
			}

			return GuildOptions{
				ID:            guildID,
				CommandPrefix: commandPrefix,
				Channels:      c,
			}
		},
	}
}
