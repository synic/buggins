package mod

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/synic/glap"

	"github.com/synic/buggins/internal/store"
)

type ConfigCommandOptions struct {
	Args       []*glap.Arg
	GetData    func(m *glap.Matches) any
	ModuleName string
	KeyArg     string
	GetKey     func(m *glap.Matches) string
}

func FetchModuleConfiguration[T any](
	ctx context.Context,
	db *store.Queries,
	module string,
) ([]T, error) {
	var configs []T

	rows, err := db.FindModuleConfigurations(ctx, module)

	if err != nil {
		return configs, err
	}

	for _, row := range rows {
		var config T

		data, ok := row.Data.([]byte)

		if !ok {
			return configs, fmt.Errorf(
				"could not parse '%s' configuration item '%s'",
				module,
				row.Key,
			)
		}

		err := json.Unmarshal(data, &config)

		if err != nil {
			return configs, fmt.Errorf(
				"could not parse featured configuration for guild %s: %w",
				row.Key,
				err,
			)
		}

		configs = append(configs, config)
	}

	return configs, nil
}
