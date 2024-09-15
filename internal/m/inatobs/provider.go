package inatobs

import (
	"github.com/charmbracelet/log"

	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

func Provider(db *store.Queries, logger *log.Logger) (m.ModuleProviderResult, error) {
	module, err := New(db, logger)

	if err != nil {
		return m.ModuleProviderResult{}, err
	}

	return m.ModuleProviderResult{Module: module}, nil
}
