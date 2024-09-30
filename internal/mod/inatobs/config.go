package inatobs

import (
	"github.com/spf13/pflag"

	"github.com/synic/buggins/internal/mod"
)

var (
	pageSize    int
	channelID   string
	projectID   int64
	cronPattern string
)

type ChannelConfig struct {
	ID          string `json:"id"`
	CronPattern string `json:"cron_pattern"`
	ProjectID   int64  `json:"inat_project_id"`
	PageSize    int    `json:"page_size"`
}

func GetConfigCommandOptions() mod.ConfigCommandOptions {
	flags := pflag.NewFlagSet(moduleName, pflag.ExitOnError)

	flags.StringVarP(&channelID, "channel-id", "c", "", "Channel ID")
	flags.Int64VarP(&projectID, "project-id", "p", 0, "iNaturalist Project ID")
	flags.StringVar(&cronPattern, "schedule-pattern", "0 * * * *", "Schedule pattern")
	flags.IntVar(&pageSize, "page-size", 10, "Number of pages to fetch from iNaturalist project")

	return mod.ConfigCommandOptions{
		Flags:         flags,
		KeyFlag:       "channel-id",
		RequiredFlags: []string{"channel-id", "project-id"},
		ModuleName:    moduleName,
		GetKey:        func() string { return channelID },
		GetData: func() any {
			return ChannelConfig{
				ID:          channelID,
				ProjectID:   projectID,
				PageSize:    pageSize,
				CronPattern: cronPattern,
			}
		},
	}
}
