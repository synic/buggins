package inatobs

import (
	"github.com/synic/glap"

	"github.com/synic/buggins/internal/mod"
)

type ChannelConfig struct {
	ID          string `json:"id"`
	CronPattern string `json:"cron_pattern"`
	ProjectID   int64  `json:"inat_project_id"`
	PageSize    int    `json:"page_size"`
}

func ConfigCommandOptions() mod.ConfigCommandOptions {
	args := []*glap.Arg{
		glap.NewArg("channel-id").Short('c').Required(true).Help("Channel CHANNEL_ID"),
		glap.NewArg("project-id").Short('p').Required(true).Help("Project PROJECT_ID"),
		glap.NewArg("schedule-pattern").Default("0 * * * *").Help("Schedule cron pattern PATTERN"),
		glap.NewArg("page-size").Default("10").Help("Number of pages to fetch from iNaturalist SIZE"),
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
			projectID, _ := m.GetInt64("project-id")
			pageSize, _ := m.GetInt("page-size")
			cronPattern, _ := m.GetString("schedule-pattern")
			return ChannelConfig{
				ID:          channelID,
				ProjectID:   projectID,
				PageSize:    pageSize,
				CronPattern: cronPattern,
			}
		},
	}
}
