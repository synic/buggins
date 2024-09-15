package featured

import (
	"context"
	"encoding/json"
	"log"

	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "featured"

	guildID               string
	channelID             string
	RequiredReactionCount int
)

type GuildOptions struct {
	ID                    string `json:"guild_id"`
	ChannelID             string `json:"channel_id"`
	RequiredReactionCount int    `json:"reaction_count"`
}

type Options struct {
	Guilds []GuildOptions
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
			log.Printf("could not parse featured configuration for guild %s", row.Key)
			continue
		}
		err := json.Unmarshal(data, &options)

		if err != nil {
			log.Printf("could not parse featured configuration for guild %s: %v", row.Key, err)
			continue
		}

		guilds = append(guilds, options)
	}

	return Options{Guilds: guilds}, nil
}

func GetConfigCommandOptions() m.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&guildID, "guild-id", "g", "", "Guild ID")
	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")
	flags.IntVarP(&RequiredReactionCount, "reaction-count", "r", 6, "Required reaction count")

	return m.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "guild-id",
		RequiredFlags: []string{"guild-id", "channel-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return guildID },
		GetData: func() any {
			return GuildOptions{
				ID:                    guildID,
				ChannelID:             channelID,
				RequiredReactionCount: RequiredReactionCount,
			}
		},
	}
}
