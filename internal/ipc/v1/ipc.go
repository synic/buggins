package ipc

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

type Service struct {
	UnimplementedIpcServiceServer
	discord *discordgo.Session
	manager *mod.ModuleManager
	db      *store.Queries
	logger  *slog.Logger
}

func New(
	discord *discordgo.Session,
	db *store.Queries,
	manager *mod.ModuleManager,
	logger *slog.Logger,

) (*Service, error) {
	return &Service{discord: discord, manager: manager, db: db, logger: logger}, nil
}

func (s *Service) ReloadConfiguration(
	ctx context.Context,
	request *ReloadConfigurationRequest,
) (*emptypb.Empty, error) {
	for _, m := range s.manager.Modules() {
		if m.Name() == request.Module {
			s.logger.Info("Reloading configuration", "module", m.Name())
			m.ReloadConfig(ctx, s.discord, s.db)
		}
	}
	return &emptypb.Empty{}, nil
}
