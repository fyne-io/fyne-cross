package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucor/fyne-cross/v2/internal/volume"
)

var flagSet = flag.NewFlagSet("fyne-cross", flag.ExitOnError)

// CommonFlags holds the flags shared between all commands
type CommonFlags struct {
	// AppID represents the application ID used for distribution
	AppID string
	// CacheDir is the directory used to share/cache sources and dependencies.
	// Default to system cache directory (i.e. $HOME/.cache/fyne-cross)
	CacheDir string
	// DockerImage represents a custom docker image to use for build
	DockerImage string
	// Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	Env envFlag
	// Icon represents the application icon used for distribution
	Icon string
	// Ldflags represents the flags to pass to the external linker
	Ldflags string
	// NoCache if true will not use the go build cache
	NoCache bool
	// NoStripDebug if true will not strip debug information from binaries
	NoStripDebug bool
	// Output represents the named output file
	Output string
	// RootDir represents the project root directory
	RootDir string
	// Silent enables the silent mode
	Silent bool
	// Debug enables the debug mode
	Debug bool
}

// newCommonFlags defines all the flags for the shared options
func newCommonFlags() (*CommonFlags, error) {
	output, err := defaultOutput()
	if err != nil {
		return nil, err
	}
	rootDir, err := volume.DefaultWorkDirHost()
	if err != nil {
		return nil, err
	}
	cacheDir, err := volume.DefaultCacheDirHost()
	if err != nil {
		return nil, err
	}

	defaultIcon, err := volume.DefaultIconHost()
	if err != nil {
		return nil, err
	}

	flags := &CommonFlags{}
	flagSet.StringVar(&flags.AppID, "app-id", output, "Application ID used for distribution")
	flagSet.StringVar(&flags.CacheDir, "cache", cacheDir, "Directory used to share/cache sources and dependencies")
	flagSet.BoolVar(&flags.NoCache, "no-cache", false, "Do not use the go build cache")
	flagSet.Var(&flags.Env, "env", "List of additional env variables specified as KEY=VALUE and separated by comma")
	flagSet.StringVar(&flags.Icon, "icon", defaultIcon, "Application icon used for distribution")
	flagSet.StringVar(&flags.DockerImage, "image", "", "Custom docker image to use for build")
	flagSet.StringVar(&flags.Ldflags, "ldflags", "", "Additional flags to pass to the external linker")
	flagSet.BoolVar(&flags.NoStripDebug, "no-strip-debug", false, "Do not strip debug information from binaries")
	flagSet.StringVar(&flags.Output, "output", output, "Named output file")
	flagSet.StringVar(&flags.RootDir, "dir", rootDir, "Fyne app root directory")
	flagSet.BoolVar(&flags.Silent, "silent", false, "Silent mode")
	flagSet.BoolVar(&flags.Debug, "debug", false, "Debug mode")
	return flags, nil
}

func defaultOutput() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get the path for current directory %s", err)
	}
	_, output := filepath.Split(wd)
	return output, nil
}

// envFlag is a custom flag used to define custom env variables
type envFlag []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ef *envFlag) String() string {
	return fmt.Sprint(*ef)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (ef *envFlag) Set(value string) error {
	*ef = []string{}
	if len(*ef) > 1 {
		return errors.New("flag already set")
	}

	for _, v := range strings.Split(value, ",") {

		*ef = append(*ef, v)
	}

	// validate env vars
	for _, v := range *ef {
		parts := strings.Split(v, "=")
		if len(parts) != 2 {
			return errors.New("env var must defined as KEY=VALUE or KEY=")
		}
	}

	return nil
}

// targetArchFlag is a custom flag used to define architectures
type targetArchFlag []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (af *targetArchFlag) String() string {
	return fmt.Sprint(*af)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (af *targetArchFlag) Set(value string) error {
	*af = []string{}
	if len(*af) > 1 {
		return errors.New("flag already set")
	}

	for _, v := range strings.Split(value, ",") {
		*af = append(*af, v)
	}
	return nil
}
