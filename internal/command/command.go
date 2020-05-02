package command

import (
	"os"

	"github.com/lucor/fyne-cross/v2/internal/log"
)

// Command wraps the methods for a fyne-cross command
type Command interface {
	Name() string              // Name returns the one word command name
	Description() string       // Description returns the command description
	Parse(args []string) error // Parse parses the cli arguments
	Usage()                    // Usage displays the command usage
	Run() error                // Run runs the command
}

// Usage prints the fyne-cross command usage
func Usage(commands []Command) {
	template := `fyne-cross is a simple tool to cross compile Fyne applications

Usage: fyne-cross <command> [options]

The commands are:

{{ range $k, $cmd := . }}	{{ printf "%-13s %s\n" $cmd.Name $cmd.Description }}{{ end }}
Use "fyne-cross <command> --help" for more information about a command.
`

	printUsage(template, commands)
}

func printUsage(template string, data interface{}) {
	log.PrintTemplate(os.Stderr, template, data)
}
