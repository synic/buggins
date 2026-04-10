package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/synic/glap"
)

var (
	databaseFile = "db.sqlite"
	subcommands  []*glap.Command
)

func RegisterCommand(cmd *glap.Command) {
	subcommands = append(subcommands, cmd)
}

func Execute() {
	app := glap.NewCommand("buggins").
		About("Discord bot for the Macromania server").
		SubcommandRequired(true).
		Arg(glap.NewArg("database-file").
			Default("db.sqlite").
			Env("DATABASE_FILE").
			Help("Database file")).
		Run(func(m *glap.Matches) error {
			if v, ok := m.GetString("database-file"); ok {
				databaseFile = v
			}
			return nil
		})

	for _, cmd := range subcommands {
		app.Subcommand(cmd)
	}

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		var helpErr *glap.HelpRequestedError
		var versionErr *glap.VersionRequestedError
		if errors.As(err, &helpErr) || errors.As(err, &versionErr) {
			fmt.Println(err)
			return
		}
		log.Fatal(err)
	}
}
