package m

import "go.uber.org/fx"

type ModuleProviderResult struct {
	fx.Out
	Module Module `group:"modules"`
}

type ModuleManagerParams struct {
	fx.In

	Modules []Module `group:"modules"`
}
