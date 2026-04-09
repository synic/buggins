package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/synic/glap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/synic/buggins/internal/ipc/v1"
	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/mod/featured"
	"github.com/synic/buggins/internal/mod/inatlookup"
	"github.com/synic/buggins/internal/mod/inatobs"
	"github.com/synic/buggins/internal/mod/thisthat"
	"github.com/synic/buggins/internal/store"
)

var (
	ipcSocket              string
	shouldConnectIpcService bool
)

var configCommandFunctions = []func() mod.ConfigCommandOptions{
	featured.ConfigCommandOptions,
	thisthat.ConfigCommandOptions,
	inatobs.ConfigCommandOptions,
	inatlookup.ConfigCommandOptions,
}

func maybeSendReload(ctx context.Context, module string) {
	socket := fmt.Sprintf("unix://%s", ipcSocket)

	if !shouldConnectIpcService {
		return
	}

	if _, err := os.Stat(ipcSocket); errors.Is(err, os.ErrNotExist) {
		logger.Debug("ipc socket not found, skipping reload signal")
		return
	}

	conn, err := grpc.NewClient(
		socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		logger.Error("error connecting to ipc server", "err", err)
		return
	}

	defer conn.Close()
	client := ipc.NewIpcServiceClient(conn)

	_, err = client.ReloadConfiguration(ctx, &ipc.ReloadConfigurationRequest{
		Module: module,
	})

	if err != nil {
		logger.Error("error sending reload signal", "err", err)
		return
	}

}

func saveConfigurationOption(c mod.ConfigCommandOptions, m *glap.Matches) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	options := c.GetData(m)
	key := c.GetKey(m)

	_, err = db.FindModuleConfiguration(
		ctx,
		store.FindModuleConfigurationParams{Module: c.ModuleName, Key: key},
	)

	if err == nil {
		return fmt.Errorf("config %s already exists", key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error fetching existing configuration: %v", err)
	}

	data, err := json.Marshal(options)

	if err != nil {
		return err
	}

	_, err = db.CreateModuleConfiguration(ctx, store.CreateModuleConfigurationParams{
		Module: c.ModuleName,
		Key:    key,
		Data:   data,
	})

	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	maybeSendReload(ctx, c.ModuleName)

	return nil
}

func updateConfigurationOption(c mod.ConfigCommandOptions, m *glap.Matches) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	key := c.GetKey(m)

	_, err = db.FindModuleConfiguration(
		ctx,
		store.FindModuleConfigurationParams{Module: c.ModuleName, Key: key},
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("config %s not found", key)
		}

		logger.Error("error looking up configuration", "err", err)
		return err
	}

	options := c.GetData(m)
	data, err := json.Marshal(options)

	if err != nil {
		return err
	}

	_, err = db.UpdateModuleConfiguration(ctx, store.UpdateModuleConfigurationParams{
		Module: c.ModuleName,
		Key:    key,
		Data:   data,
	})

	if err != nil {
		logger.Error("could not delete configuration", "err", err)
		return err
	}

	maybeSendReload(ctx, c.ModuleName)

	return nil
}

func removeConfigurationOption(c mod.ConfigCommandOptions, m *glap.Matches) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	key := c.GetKey(m)

	_, err = db.FindModuleConfiguration(
		ctx,
		store.FindModuleConfigurationParams{Module: c.ModuleName, Key: key},
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("config %s not found", key)
		}

		logger.Error("error looking up configuration", "err", err)
		return err
	}

	_, err = db.DeleteModuleConfiguration(ctx, store.DeleteModuleConfigurationParams{
		Module: c.ModuleName,
		Key:    key,
	})

	if err != nil {
		logger.Error("could not delete configuration", "err", err)
		return err
	}

	maybeSendReload(ctx, c.ModuleName)
	return nil
}

func init() {
	configCmd := glap.NewCommand("config").
		About("Configure a module").
		Arg(glap.NewArg("ipc-socket").
			Default("/tmp/buggins-ipc.sock").
			Help("IPC socket location")).
		Arg(glap.NewArg("connect-ipc").
			Action(glap.SetTrue).
			Default("true").
			Help("Attempt to use IPC to reload module configuration")).
		Run(func(m *glap.Matches) error {
			if v, ok := m.GetString("ipc-socket"); ok {
				ipcSocket = v
			}
			if v, ok := m.GetBool("connect-ipc"); ok {
				shouldConnectIpcService = v
			}
			return nil
		})

	for _, f := range configCommandFunctions {
		c := f()

		modCmd := glap.NewCommand(c.ModuleName).
			About(fmt.Sprintf("Configure module '%s'", c.ModuleName))

		addCmd := glap.NewCommand("add").
			About("Add a configuration").
			Run(func(m *glap.Matches) error {
				err := saveConfigurationOption(c, m)
				if err != nil {
					logger.Error("error adding config", "key", c.GetKey(m), "err", err)
					return err
				}
				logger.Info("Configuration added successfully!")
				return nil
			})

		updateCmd := glap.NewCommand("update").
			About("Update a configuration").
			Run(func(m *glap.Matches) error {
				err := updateConfigurationOption(c, m)
				if err != nil {
					logger.Error("error updating config", "key", c.GetKey(m), "err", err)
					return err
				}
				logger.Info("Configuration updated successfully!")
				return nil
			})

		rmCmd := glap.NewCommand("rm").
			About("Remove a configuration").
			Run(func(m *glap.Matches) error {
				err := removeConfigurationOption(c, m)
				if err != nil {
					logger.Error("error removing config", "key", c.GetKey(m), "err", err)
					return err
				}
				logger.Info("Configuration removed.")
				return nil
			})

		for _, arg := range c.Args {
			addCmd.Arg(arg.Clone())
			updateCmd.Arg(arg.Clone())
			if arg.GetName() == c.KeyArg {
				rmCmd.Arg(arg.Clone())
			}
		}

		modCmd.Subcommand(addCmd).Subcommand(updateCmd).Subcommand(rmCmd)
		configCmd.Subcommand(modCmd)
	}

	RegisterCommand(configCmd)
}
