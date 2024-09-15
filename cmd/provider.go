package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/synic/buggins/internal/ipc/v1"
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/m/featured"
	"github.com/synic/buggins/internal/m/inatlookup"
	"github.com/synic/buggins/internal/m/inatobs"
	"github.com/synic/buggins/internal/m/thisthat"
	"github.com/synic/buggins/internal/store"
)

func getProviders(databaseURL string) fx.Option {
	return fx.Options(
		fx.Provide(newDatabase(databaseURL)),
		fx.Provide(featured.Provider),
		fx.Provide(inatobs.Provider),
		fx.Provide(inatlookup.Provider),
		fx.Provide(thisthat.Provider),
		fx.Provide(newModuleManager),
	)
}

func newDiscordSession(
	token string,
) func(fx.Lifecycle, *m.ModuleManager) (*discordgo.Session, error) {
	return func(lc fx.Lifecycle, bot *m.ModuleManager) (*discordgo.Session, error) {
		discord, err := discordgo.New(fmt.Sprintf("Bot %s", token))

		if err != nil {
			return nil, err
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
					log.Printf("User '%s' connected to discord!", r.User.Username)

					for _, module := range bot.Modules() {
						module.Start(discord)
					}
				})

				if err := discord.Open(); err != nil {
					return err
				}

				log.Println("started discord bot")

				return nil
			},
			OnStop: func(ctx context.Context) error {
				log.Println("closing discord connection...")
				if err := discord.Close(); err != nil {
					return err
				}

				return nil
			},
		})

		return discord, nil
	}
}

func newModuleManager(params m.ModuleManagerParams) (*m.ModuleManager, error) {
	return m.NewManager(params.Modules)
}

func newDatabase(url string) func() (*store.Queries, error) {
	return func() (*store.Queries, error) {
		return store.Init(url)
	}
}

func startIpcService(
	bind string,
) func(
	fx.Lifecycle,
	*discordgo.Session,
	*m.ModuleManager,
	*store.Queries,
) (*ipc.Service, error) {
	return func(
		lc fx.Lifecycle,
		discord *discordgo.Session,
		manager *m.ModuleManager,
		db *store.Queries,
	) (*ipc.Service, error) {
		var opts []grpc.ServerOption
		service, err := ipc.New(discord, db, manager)
		if err != nil {
			return nil, err
		}

		lis, err := net.Listen("unix", bind)

		if err != nil {
			return nil, err
		}

		grpcServer := grpc.NewServer(opts...)
		ipc.RegisterIpcServiceServer(grpcServer, service)

		log.Printf("ipc service serving on %s", bind)

		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go grpcServer.Serve(lis)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				log.Println("stopping IPC service...")
				grpcServer.Stop()
				lis.Close()
				return nil
			},
		})

		return service, nil
	}
}

func provideIpcService(
	bind string,
) fx.Option {
	if bind == "" {
		return fx.Options()
	}

	return fx.Options(
		fx.Provide(startIpcService(bind)),
		fx.Invoke(func(*ipc.Service) {}),
	)
}
