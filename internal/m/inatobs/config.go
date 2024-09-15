package inatobs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

var (
	moduleName = "inatobs"

	pageSize    int
	channelID   string
	projectID   int64
	cronPattern string
)

type ChannelOptions struct {
	ID          string `json:"id"`
	CronPattern string `json:"cron_pattern"`
	ProjectID   int64  `json:"inat_project_id"`
	PageSize    int    `json:"page_size"`
}

type Options struct {
	Channels []ChannelOptions `json:"channels"`
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
			log.Printf("could not parse inatobs configuration for guild %s", row.Key)
			continue
		}
		err := json.Unmarshal(data, &options)

		if err != nil {
			log.Printf("could not parse inatobs configuration for guild %s: %v", row.Key, err)
			continue
		}

		channels = append(channels, options)
	}

	return Options{Channels: channels}, nil
}

func GetConfigCommandOptions() m.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")
	flags.Int64VarP(&projectID, "project-id", "p", 0, "iNaturalist Project ID")
	flags.StringVar(&cronPattern, "schedule-pattern", "0 * * * *", "Schedule pattern")
	flags.IntVar(&pageSize, "page-size", 10, "Number of pages to fetch from iNaturalist project")

	return m.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "channel-id",
		RequiredFlags: []string{"channel-id", "project-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return channelID },
		GetData: func() any {
			return ChannelOptions{
				ID:          channelID,
				ProjectID:   projectID,
				PageSize:    pageSize,
				CronPattern: cronPattern,
			}
		},
	}
}
