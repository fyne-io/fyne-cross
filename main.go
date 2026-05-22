package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/command"
	"github.com/fyne-io/fyne-cross/internal/log"
)

func main() {
	app := &cli.App{
		Name:        "fyne-cross",
		Usage:       "A simple tool to cross compile Fyne applications.",
		HideVersion: true,
		Commands: []*cli.Command{
			command.DarwinSDKExtract(),
			command.Darwin(),
			command.Linux(),
			command.Windows(),
			command.Android(),
			command.IOS(),
			command.FreeBSD(),
			command.Web(),
			command.Version(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("[✗] %s", err)
	}
}
