package cmd

import (
	"github.com/synic/glap"

	"github.com/synic/buggins/internal/store"
)

func init() {
	cmd := glap.NewCommand("migrate").
		About("Run database migrations and exit").
		Run(func(m *glap.Matches) error {
			if err := store.Migrate(); err != nil {
				logger.Error("error running migrations", "err", err)
				return err
			}
			logger.Info("Migrations complete")
			return nil
		})

	RegisterCommand(cmd)
}
