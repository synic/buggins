package featured

import (
	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/mod"
)

var (
	guildID               string
	channelID             string
	RequiredReactionCount int
)

type GuildConfig struct {
	ID                    string `json:"guild_id"`
	ChannelID             string `json:"channel_id"`
	RequiredReactionCount int    `json:"reaction_count"`
}

func GetConfigCommandOptions() mod.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&guildID, "guild-id", "g", "", "Guild ID")
	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")
	flags.IntVarP(&RequiredReactionCount, "reaction-count", "r", 6, "Required reaction count")

	return mod.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "guild-id",
		RequiredFlags: []string{"guild-id", "channel-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return guildID },
		GetData: func() any {
			return GuildConfig{
				ID:                    guildID,
				ChannelID:             channelID,
				RequiredReactionCount: RequiredReactionCount,
			}
		},
	}
}
