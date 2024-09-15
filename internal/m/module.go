package m

import (
	"github.com/bwmarrin/discordgo"

	"github.com/synic/buggins/internal/store"
)

type Module interface {
	Start(*discordgo.Session) error
	ReloadConfig(*discordgo.Session, *store.Queries) error
	Name() string
}
