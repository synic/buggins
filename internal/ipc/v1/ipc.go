package ipc

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

type Service struct {
	UnimplementedIpcServiceServer
	discord *discordgo.Session
	manager *m.ModuleManager
	db      *store.Queries
}

func New(
	discord *discordgo.Session,
	db *store.Queries,
	manager *m.ModuleManager,
) (*Service, error) {
	return &Service{discord: discord, manager: manager, db: db}, nil
}

func (s *Service) ReloadConfiguration(
	ctx context.Context,
	request *ReloadConfigurationRequest,
) (*emptypb.Empty, error) {
	for _, m := range s.manager.Modules() {
		if m.Name() == request.Module {
			log.Printf("Reloading configuration for module '%s'...", m.Name())
			m.ReloadConfig(s.discord, s.db)
		}
	}
	return &emptypb.Empty{}, nil
}
