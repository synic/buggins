package mod

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"github.com/synic/buggins/internal/store"
)

type Module interface {
	Start(context.Context, *discordgo.Session, *store.Queries) error
	ReloadConfig(context.Context, *discordgo.Session, *store.Queries) error
	Name() string
}

type ModuleManager struct {
	modules     []Module
	modulesLock sync.RWMutex
}

func NewManager(modules []Module) (*ModuleManager, error) {
	return &ModuleManager{modules: modules}, nil
}

func (m *ModuleManager) Modules() []Module {
	m.modulesLock.RLock()
	defer m.modulesLock.RUnlock()
	return m.modules
}

func (m *ModuleManager) SetModules(modules []Module) {
	m.modulesLock.Lock()
	defer m.modulesLock.Unlock()
	m.modules = modules
}

func Provider(params ModuleManagerParams) (*ModuleManager, error) {
	return NewManager(params.Modules)
}

type ModuleProviderResult struct {
	fx.Out
	Module Module `group:"modules"`
}

type ModuleManagerParams struct {
	fx.In

	Modules []Module `group:"modules"`
}
