package thisthat

import (
	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/mod"
)

var (
	channelID string
)

type ChannelConfig struct {
	ID string `json:"id"`
}

func GetConfigCommandOptions() mod.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")

	return mod.ConfigCommandOptions{
		ModuleName:    moduleName,
		KeyFlag:       "channel-id",
		Flags:         flags,
		GetKey:        func() string { return channelID },
		RequiredFlags: []string{"channel-id"},
		GetData: func() any {
			return ChannelConfig{
				ID: channelID,
			}
		},
	}
}
