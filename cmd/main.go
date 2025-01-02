package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const defaultServerAddr = ":8080"
const defaultAPIURL = "http://localhost:8080"

func main() {
	// setup cli app
	app := &cli.App{
		Name:  "stagerunner",
		Usage: "A pipeline execution service",
		Commands: []*cli.Command{
			serverCommand,
			clientCommand,
		},
	}

	// run cli app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
