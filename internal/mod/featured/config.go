package featured

import (
	"github.com/synic/glap"

	"github.com/synic/buggins/internal/mod"
)

type GuildConfig struct {
	ID                    string `json:"guild_id"`
	ChannelID             string `json:"channel_id"`
	RequiredReactionCount int    `json:"reaction_count"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	args := []*glap.Arg{
		glap.NewArg("guild-id").Short('g').Required(true).Help("Guild GUILD_ID"),
		glap.NewArg("channel-id").Short('c').Required(true).Help("Channel CHANNEL_ID"),
		glap.NewArg("reaction-count").Short('r').Default("6").Help("Number of reactions to trigger COUNT"),
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
			channelID, _ := m.GetString("channel-id")
			reactionCount, _ := m.GetInt("reaction-count")
			return GuildConfig{
				ID:                    guildID,
				ChannelID:             channelID,
				RequiredReactionCount: reactionCount,
			}
		},
	}
}
