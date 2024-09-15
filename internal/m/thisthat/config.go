package thisthat

import (
	"context"
	"encoding/json"
	"log"

	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "thisthat"
	channelID  string
)

type ChannelOptions struct {
	ID string `json:"id"`
}

type Options struct {
	Channels []ChannelOptions
}

func fetchModuleOptions(db *store.Queries) (Options, error) {
	ctx := context.Background()
	rows, err := db.FindModuleConfigurations(ctx, moduleName)

	if err != nil {
		return Options{}, err
	}

	channels := make([]ChannelOptions, 0, len(rows))

	for _, row := range rows {
		var options ChannelOptions

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

		channels = append(channels, options)
	}

	return Options{Channels: channels}, nil
}

func GetConfigCommandOptions() m.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")

	return m.ConfigCommandOptions{
		ModuleName:    moduleName,
		KeyFlag:       "channel-id",
		Flags:         flags,
		GetKey:        func() string { return channelID },
		RequiredFlags: []string{"channel-id"},
		GetData: func() any {
			return ChannelOptions{
				ID: channelID,
			}
		},
	}
}
