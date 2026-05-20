package main

import (
	//"fmt"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/command"
	"github.com/fyne-io/fyne-cross/internal/log"
)

func main() {
	//flags := &commands.CommonFlags{}

	app := &cli.App{
		Name:  "fyne-cross",
		Usage: "A simple tool to cross compile Fyne applications.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "app-build",
				//Destination: &flags.AppBuild,
			},
		},
		Commands: []*cli.Command{
			command.DarwinSDKExtract(),
/*
			commands.Darwin(),
			commands.Linux(),
			commands.Windows(),
			commands.Android(),
			commands.IOS(),
			commands.FreeBSD(),
			commands.Web(),
*/
			command.Version(),
		},
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		app.Version = info.Main.Version
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("[✗] %s", err)
	}
}
