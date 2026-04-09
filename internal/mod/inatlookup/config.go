package inatlookup

import (
	"regexp"
	"slices"

	"github.com/synic/glap"

	"github.com/synic/buggins/internal/mod"
)

type GuildConfig struct {
	CommandPrefixRegex *regexp.Regexp `json:"-"`
	Name               string         `json:"name"`
	ID                 string         `json:"id"`
	CommandPrefix      string         `json:"command_prefix"`
	Channels           []string       `json:"channels"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	args := []*glap.Arg{
		glap.NewArg("guild-id").Short('g').Required(true).Help("Guild GUILD_ID"),
		glap.NewArg("command-prefix").Default(",").Help("Command prefix PREFIX"),
		glap.NewArg("channels").Short('c').Action(glap.Append).Help("Channel ids (omit for all channels) CHANNEL_IDS"),
	}

	return mod.ConfigCommandOptions{
		Args:       args,
		KeyArg:     "guild-id",
		ModuleName: moduleName,
		GetKey: func(m *glap.Matches) string {
			v, _ := m.GetString("guild-id")
			return v
		},
		GetData: func(m *glap.Matches) any {
			guildID, _ := m.GetString("guild-id")
			commandPrefix, _ := m.GetString("command-prefix")
			channels, _ := m.GetStringSlice("channels")

			if slices.Contains(channels, "all") {
				channels = []string{}
			}

			return GuildConfig{
				ID:            guildID,
				CommandPrefix: commandPrefix,
				Channels:      channels,
			}
		},
	}
}
