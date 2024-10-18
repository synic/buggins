package inatobs

import (
	"github.com/urfave/cli/v2"

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

func ConfigCommandOptions() mod.ConfigCommandOptions {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "channel-id",
			Usage:       "Channel ID",
			Aliases:     []string{"c"},
			Destination: &channelID,
			Required:    true,
		},
		&cli.Int64Flag{
			Name:        "project-id",
			Usage:       "Project ID",
			Aliases:     []string{"p"},
			Destination: &projectID,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "schedule-pattern",
			Destination: &cronPattern,
			Value:       "0 * * * *",
			Usage:       "Schedule pattern",
		},
		&cli.IntFlag{
			Name:        "page-size",
			Value:       10,
			Usage:       "Number of pages to fetch from iNaturalist",
			Destination: &pageSize,
		},
	}

	return mod.ConfigCommandOptions{
		Flags:      flags,
		KeyFlag:    "channel-id",
		ModuleName: moduleName,
		GetKey:     func() string { return channelID },
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
