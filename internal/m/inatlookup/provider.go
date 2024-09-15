package inatlookup

import (
	"github.com/synic/buggins/internal/m"
	"github.com/synic/buggins/internal/store"
)

func Provider(db *store.Queries) (m.ModuleProviderResult, error) {
	module, err := New(db)

	if err != nil {
		return m.ModuleProviderResult{}, err
	}

	return m.ModuleProviderResult{Module: module}, nil
}
