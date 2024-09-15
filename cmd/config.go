package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/synic/buggins/internal/ipc/v1"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/m/featured"
	"github.com/synic/buggins/internal/m/inatlookup"
	"github.com/synic/buggins/internal/m/inatobs"
	"github.com/synic/buggins/internal/m/thisthat"
	"github.com/synic/buggins/internal/store"
)

var (
	shouldConnectIpcService bool
)

var configCommandFunctions = []func() m.ConfigCommandOptions{
	featured.GetConfigCommandOptions,
	thisthat.GetConfigCommandOptions,
	inatobs.GetConfigCommandOptions,
	inatlookup.GetConfigCommandOptions,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure a module",
}

func maybeSendReload(ctx context.Context, module string) {
	if !shouldConnectIpcService {
		return
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("unix://%s", ipcSocket),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Printf("error connecting to ipc server: %v", err)
		return
	}

	defer conn.Close()
	client := ipc.NewIpcServiceClient(conn)

	_, err = client.ReloadConfiguration(ctx, &ipc.ReloadConfigurationRequest{
		Module: module,
	})

	if err != nil {
		log.Printf("error sending reload signal: %v", err)
		return
	}

}

func saveConfigurationOption(c m.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(viper.GetString("DatabaseURL"))

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
		Module:  c.ModuleName,
		Key:     key,
		Options: data,
	})

	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	maybeSendReload(ctx, c.ModuleName)

	return nil
}

func updateConfigurationOption(c m.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(viper.GetString("DatabaseURL"))

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

		log.Fatalf("error looking up configuration: %v", err)
	}

	options := c.GetData()
	data, err := json.Marshal(options)

	if err != nil {
		return err
	}

	_, err = db.UpdateModuleConfiguration(ctx, store.UpdateModuleConfigurationParams{
		Module:  c.ModuleName,
		Key:     key,
		Options: data,
	})

	if err != nil {
		log.Fatalf("could not delete configuration: %v", err)
	}

	maybeSendReload(ctx, c.ModuleName)

	return nil
}

func removeConfigurationOption(c m.ConfigCommandOptions) error {
	ctx := context.Background()
	db, err := store.Init(viper.GetString("DatabaseURL"))

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

		log.Fatalf("error looking up configuration: %v", err)
	}

	_, err = db.DeleteModuleConfiguration(ctx, store.DeleteModuleConfigurationParams{
		Module: c.ModuleName,
		Key:    key,
	})

	if err != nil {
		log.Fatalf("could not delete configuration: %v", err)
	}

	maybeSendReload(ctx, c.ModuleName)
	return nil
}

func init() {
	for _, f := range configCommandFunctions {
		c := f()

		modCmd := &cobra.Command{
			Use:   c.ModuleName,
			Short: fmt.Sprintf("Configure module '%s'", c.ModuleName),
		}

		addCmd := &cobra.Command{
			Use:   "add",
			Short: "Add a configuration",
			Run: func(*cobra.Command, []string) {
				err := saveConfigurationOption(c)

				if err != nil {
					log.Fatalf("error adding config for %s: %v", c.GetKey(), err)
				}

				log.Println("Configuration added successfully!")
			},
		}

		updateCmd := &cobra.Command{
			Use:   "update",
			Short: "Update a configuration",
			Run: func(*cobra.Command, []string) {
				err := updateConfigurationOption(c)

				if err != nil {
					log.Fatalf("error updating config for %s: %v", c.GetKey(), err)
				}

				log.Println("Configuration updated successfully!")
			},
		}

		rmCmd := &cobra.Command{
			Use:   "rm",
			Short: "Remove a configuration",
			Run: func(*cobra.Command, []string) {
				err := removeConfigurationOption(c)

				if err != nil {
					log.Fatalf("error removing config for %s: %v", c.GetKey(), err)
				}

				log.Println("Configuration removed.")
			},
		}

		c.Flags.VisitAll(func(f *pflag.Flag) {
			addCmd.Flags().AddFlag(f)
			updateCmd.Flags().AddFlag(f)
			if slices.Contains(c.RequiredFlags, f.Name) {
				addCmd.MarkFlagRequired(f.Name)
				updateCmd.MarkFlagRequired(f.Name)
			}

			if f.Name == c.KeyFlag {
				rmCmd.Flags().AddFlag(f)
				rmCmd.MarkFlagRequired(f.Name)
			}
		})

		modCmd.AddCommand(addCmd)
		modCmd.AddCommand(updateCmd)
		modCmd.AddCommand(rmCmd)
		configCmd.AddCommand(modCmd)
	}

	configCmd.PersistentFlags().
		StringVar(&ipcSocket, "ipc-socket", "/tmp/buggins-ipc.sock", "IPC bind")
	configCmd.PersistentFlags().
		BoolVar(&shouldConnectIpcService, "connect-ipc", true, "Attempt to use IPC to reload module configuration")

	rootCmd.AddCommand(configCmd)
}
