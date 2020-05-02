package command

import (
	"fmt"
	"runtime/debug"
)

const version = "develop"

// Version is the version command
type Version struct{}

// Name returns the one word command name
func (cmd *Version) Name() string {
	return "version"
}

// Description returns the command description
func (cmd *Version) Description() string {
	return "Print the fyne-cross version information"
}

// Run runs the command
func (cmd *Version) Run() error {
	fmt.Printf("fyne-cross version %s\n", getVersion())
	return nil
}

// Parse parses the arguments and set the usage for the command
func (cmd *Version) Parse(args []string) error {
	return nil
}

// Usage displays the command usage
func (cmd *Version) Usage() {
	template := `Usage: fyne-cross version

{{ . }}
`
	printUsage(template, cmd.Description())
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return version
}
