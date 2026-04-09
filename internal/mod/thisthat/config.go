package thisthat

import (
	"github.com/synic/glap"

	"github.com/synic/buggins/internal/mod"
)

type ChannelConfig struct {
	ID string `json:"id"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	args := []*glap.Arg{
		glap.NewArg("channel-id").Short('c').Required(true).Help("Channel CHANNEL_ID"),
	}

	return mod.ConfigCommandOptions{
		Args:       args,
		KeyArg:     "channel-id",
		ModuleName: moduleName,
		GetKey: func(m *glap.Matches) string {
			v, _ := m.GetString("channel-id")
			return v
		},
		GetData: func(m *glap.Matches) any {
			channelID, _ := m.GetString("channel-id")
			return ChannelConfig{
				ID: channelID,
			}
		},
	}
}
