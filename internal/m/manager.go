package m

import "sync"

type ModuleManager struct {
	modules []Module
	mu      sync.RWMutex
}

func NewManager(modules []Module) (*ModuleManager, error) {
	return &ModuleManager{modules: modules}, nil
}

func (m *ModuleManager) Modules() []Module {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.modules
}

func (m *ModuleManager) SetModules(modules []Module) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modules = modules
}
