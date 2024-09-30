package ipc

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/synic/buggins/internal/mod"
	"github.com/synic/buggins/internal/store"
)

type Service struct {
	UnimplementedIpcServiceServer
	discord *discordgo.Session
	manager *mod.ModuleManager
	db      *store.Queries
	logger  *log.Logger
}

func New(
	discord *discordgo.Session,
	db *store.Queries,
	manager *mod.ModuleManager,
	logger *log.Logger,

) (*Service, error) {
	return &Service{discord: discord, manager: manager, db: db, logger: logger}, nil
}

func (s *Service) ReloadConfiguration(
	ctx context.Context,
	request *ReloadConfigurationRequest,
) (*emptypb.Empty, error) {
	for _, m := range s.manager.Modules() {
		if m.Name() == request.Module {
			s.logger.Infof("Reloading configuration for module '%s'...", m.Name())
			m.ReloadConfig(ctx, s.discord, s.db)
		}
	}
	return &emptypb.Empty{}, nil
}
