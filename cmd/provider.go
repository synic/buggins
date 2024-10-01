package cmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/synic/buggins/internal/ipc/v1"
	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/mod/featured"
	"github.com/synic/buggins/internal/mod/inatlookup"
	"github.com/synic/buggins/internal/mod/inatobs"
	"github.com/synic/buggins/internal/mod/thisthat"
	"github.com/synic/buggins/internal/store"
)

var (
	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
)

func getProviders(databaseFile string) fx.Option {
	return fx.Options(
		fx.Provide(newLogger),
		fx.Provide(newDatabase(databaseFile)),
		fx.Provide(featured.Provider),
		fx.Provide(inatobs.Provider),
		fx.Provide(inatlookup.Provider),
		fx.Provide(thisthat.Provider),
		fx.Provide(mod.Provider),
	)
}

func newLogger() *log.Logger {
	return logger
}

type discordSessionParams struct {
	fx.In

	LC      fx.Lifecycle
	Manager *mod.ModuleManager
	DB      *store.Queries
}

func newDiscordSession(
	token string,
) func(params discordSessionParams) (*discordgo.Session, error) {
	return func(params discordSessionParams) (*discordgo.Session, error) {
		discord, err := discordgo.New(fmt.Sprintf("Bot %s", token))

		if err != nil {
			return nil, err
		}

		params.LC.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
					logger.Infof("User '%s' connected to discord!", r.User.Username)

					for _, module := range params.Manager.Modules() {
						module.Start(ctx, discord, params.DB)
					}
				})

				if err := discord.Open(); err != nil {
					return err
				}

				logger.Info("started discord bot")

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("closing discord connection...")
				if err := discord.Close(); err != nil {
					return err
				}

				return nil
			},
		})

		return discord, nil
	}
}

func newDatabase(fileLocation string) func() (*store.Queries, error) {
	return func() (*store.Queries, error) {
		return store.Init(fileLocation)
	}
}

type ipcServiceParams struct {
	fx.In

	LC      fx.Lifecycle
	Manager *mod.ModuleManager
	DB      *store.Queries
	Discord *discordgo.Session
	Logger  *log.Logger
}

func startIpcService(bind string) func(
	params ipcServiceParams,
) (*ipc.Service, error) {
	return func(params ipcServiceParams) (*ipc.Service, error) {
		var opts []grpc.ServerOption
		service, err := ipc.New(params.Discord, params.DB, params.Manager, logger)
		if err != nil {
			return nil, err
		}

		lis, err := net.Listen("unix", bind)

		if err != nil {
			return nil, err
		}

		grpcServer := grpc.NewServer(opts...)
		ipc.RegisterIpcServiceServer(grpcServer, service)

		logger.Infof("ipc service serving on %s", bind)

		params.LC.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go grpcServer.Serve(lis)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("stopping IPC service...")
				grpcServer.Stop()
				lis.Close()
				return nil
			},
		})

		return service, nil
	}
}

func provideIpcService(bind string) fx.Option {
	if bind == "" {
		return fx.Options()
	}

	return fx.Options(
		fx.Provide(startIpcService(bind)),
		fx.Invoke(func(*ipc.Service) {}),
	)
}
