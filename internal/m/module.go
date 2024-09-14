package m

import "github.com/synic/buggins/internal/store"

type Module interface {
	Start()
	ReloadConfig(*store.Queries) error
	GetName() string
}
