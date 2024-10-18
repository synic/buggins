package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
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
	shouldConnectIpcService bool
)

var configCommandFunctions = []func() mod.ConfigCommandOptions{
	featured.ConfigCommandOptions,
	thisthat.ConfigCommandOptions,
	inatobs.ConfigCommandOptions,
	inatlookup.ConfigCommandOptions,
}

var configCmd = &cli.Command{
	Name:        "config",
	Usage:       "Configure a module",
	Subcommands: []*cli.Command{},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "ipc-socket",
			Value:       "/tmp/buggins-ipc.sock",
			Usage:       "IPC socket location",
			Destination: &ipcSocket,
		},
		&cli.BoolFlag{
			Name:        "connect-ipc",
			Value:       true,
			Usage:       "Attempt to use IPC to reload module configuration",
			Destination: &shouldConnectIpcService,
		},
	},
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

func saveConfigurationOption(c mod.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	options := c.GetData()
	key := c.GetKey()

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

func updateConfigurationOption(c mod.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	key := c.GetKey()

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

	options := c.GetData()
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

func removeConfigurationOption(c mod.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(databaseFile)

	if err != nil {
		return err
	}

	key := c.GetKey()

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
	for _, f := range configCommandFunctions {
		c := f()

		modCmd := &cli.Command{
			Name:        c.ModuleName,
			Usage:       fmt.Sprintf("Configure module '%s'", c.ModuleName),
			Subcommands: []*cli.Command{},
		}

		addCmd := &cli.Command{
			Name:  "add",
			Usage: "Add a configuration",
			Flags: []cli.Flag{},
			Action: func(*cli.Context) error {
				err := saveConfigurationOption(c)

				if err != nil {
					logger.Error("error adding config", "key", c.GetKey(), "err", err)
					return err
				}

				logger.Info("Configuration added successfully!")
				return nil
			},
		}

		updateCmd := &cli.Command{
			Name:  "update",
			Usage: "Update a configuration",
			Flags: []cli.Flag{},
			Action: func(*cli.Context) error {
				err := updateConfigurationOption(c)

				if err != nil {
					logger.Error("error updating config", "key", c.GetKey(), "err", err)
					return err
				}

				logger.Info("Configuration updated successfully!")
				return nil
			},
		}

		rmCmd := &cli.Command{
			Name:  "rm",
			Usage: "Remove a configuration",
			Action: func(*cli.Context) error {
				err := removeConfigurationOption(c)

				if err != nil {
					logger.Error("error removing config", "key", c.GetKey(), "err", err)
					return err
				}

				logger.Info("Configuration removed.")
				return nil
			},
		}

		for _, flag := range c.Flags {
			addCmd.Flags = append(addCmd.Flags, flag)
			updateCmd.Flags = append(updateCmd.Flags, flag)
		}

		modCmd.Subcommands = append(modCmd.Subcommands, addCmd, updateCmd, rmCmd)
		configCmd.Subcommands = append(configCmd.Subcommands, modCmd)
	}

	app.Commands = append(app.Commands, configCmd)
}
