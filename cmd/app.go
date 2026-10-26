package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/synic/glap"
)

var (
	databaseUrl string
	subcommands []*glap.Command
)

func RegisterCommand(cmd *glap.Command) {
	subcommands = append(subcommands, cmd)
}

func Execute() {
	app := glap.NewCommand("buggins").
		About("Discord bot for the Macromania server").
		SubcommandRequired(true).
		Arg(glap.NewArg("database-url").
			Env("DATABASE_URL").
			Help("PostgreSQL connection string")).
		Run(func(m *glap.Matches) error {
			if v, ok := m.GetString("database-url"); ok {
				databaseUrl = v
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
