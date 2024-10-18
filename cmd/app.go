package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	databaseFile = "db.sqlite"
)

var app = &cli.App{
	Name:     "buggins",
	Usage:    "Discord bot for the Macromania server",
	Commands: []*cli.Command{},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "database-file",
			Value:       "db.sqlite",
			Destination: &databaseFile,
			Usage:       "Database file",
			EnvVars:     []string{"DATABASE_FILE"},
		},
	},
}

func Execute() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
