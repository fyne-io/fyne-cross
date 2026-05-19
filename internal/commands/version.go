package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Version returns the cli command for the program version.
func Version() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print the fyne-cross version information",
		Action: func(ctx *cli.Context) error {
			fmt.Println("fyne-cross version:", ctx.App.Version)
			return nil
		},
	}
}
