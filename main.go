package main

import (
	"log"
	"os"

	"github.com/synic/buggins/cmd"
)

func main() {
	log.SetOutput(os.Stdout)
	cmd.Execute()
}
