package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/metadata"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

var flagSet = flag.NewFlagSet("fyne-cross", flag.ExitOnError)

// CommonFlags holds the flags shared between all commands
type CommonFlags struct {
	// AppBuild represents the build number, should be greater than 0 and
	// incremented for each build
	AppBuild int
	// AppID represents the application ID used for distribution
	AppID string
	// AppVersion represents the version number in the form x, x.y or x.y.z semantic version
	AppVersion string
	// CacheDir is the directory used to share/cache sources and dependencies.
	// Default to system cache directory (i.e. $HOME/.cache/fyne-cross)
	CacheDir string
	// DockerImage represents a custom docker image to use for build
	DockerImage string
	// Engine is the container engine to use
	Engine engineFlag
	// Namespace used by Kubernetes engine to run its pod in
	Namespace string
	// Base S3 directory to push and pull data from
	S3Path string
	// Container mount point size limits honored by Kubernetes only
	SizeLimit string
	// Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	Env envFlag
	// Icon represents the application icon used for distribution
	Icon string
	// Ldflags represents the flags to pass to the external linker
	Ldflags string
	// Additional build tags
	Tags tagsFlag
	// Metadata contain custom metadata passed to fyne package
	Metadata multiFlags
	// NoCache if true will not use the go build cache
	NoCache bool
	// NoProjectUpload if true, the build will be done with the artifact already stored on S3
	NoProjectUpload bool
	// NoResultDownload if true, it will leave the result of the build on S3 and won't download it locally (engine: kubernetes)
	NoResultDownload bool
	// NoStripDebug if true will not strip debug information from binaries
	NoStripDebug bool
	// NoNetwork if true will not setup network inside the container
	NoNetwork bool
	// Name represents the application name
	Name string
	// Release represents if the package should be prepared for release (disable debug etc)
	Release bool
	// RootDir represents the project root directory
	RootDir string
	// Silent enables the silent mode
	Silent bool
	// Debug enables the debug mode
	Debug bool
	// Pull attempts to pull a newer version of the docker image
	Pull bool
	// DockerRegistry changes the pull/push docker registry (defualt docker.io)
	DockerRegistry string
	//Extra mount
	ExtraMount string
}

// newCommonFlags defines all the flags for the shared options
func newCommonFlags() (*CommonFlags, error) {
	name, err := defaultName()
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

	defaultIcon := icon.Default
	appID := ""
	appVersion := "1.0.0"
	appBuild := 1

	data, _ := metadata.LoadStandard(rootDir)
	if data != nil {
		if data.Details.Icon != "" {
			defaultIcon = data.Details.Icon
		}
		if data.Details.Name != "" {
			name = data.Details.Name
		}
		if data.Details.ID != "" {
			appID = data.Details.ID
		}
		if data.Details.Version != "" {
			appVersion = data.Details.Version
		}
		if data.Details.Build != 0 {
			appBuild = data.Details.Build
		}
	}

	flags := &CommonFlags{}
	kubernetesFlagSet(flagSet, flags)
	flagSet.IntVar(&flags.AppBuild, "app-build", appBuild, "Build number, should be greater than 0 and incremented for each build")
	flagSet.StringVar(&flags.AppID, "app-id", appID, "Application ID used for distribution")
	flagSet.StringVar(&flags.AppVersion, "app-version", appVersion, "Version number in the form x, x.y or x.y.z semantic version")
	flagSet.StringVar(&flags.CacheDir, "cache", cacheDir, "Directory used to share/cache sources and dependencies")
	flagSet.BoolVar(&flags.NoCache, "no-cache", false, "Do not use the go build cache")
	flagSet.Var(&flags.Engine, "engine", "The container engine to use. Supported engines: [docker, podman, kubernetes]. Default to autodetect.")
	flagSet.Var(&flags.Env, "env", "List of additional env variables specified as KEY=VALUE")
	flagSet.StringVar(&flags.Icon, "icon", defaultIcon, "Application icon used for distribution")
	flagSet.StringVar(&flags.DockerImage, "image", "", "Custom docker image to use for build")
	flagSet.StringVar(&flags.Ldflags, "ldflags", "", "Additional flags to pass to the external linker")
	flagSet.Var(&flags.Tags, "tags", "List of additional build tags separated by comma")
	flagSet.Var(&flags.Metadata, "metadata", "Additional metadata `key=value` passed to fyne package")
	flagSet.BoolVar(&flags.NoStripDebug, "no-strip-debug", false, "Do not strip debug information from binaries")
	flagSet.StringVar(&flags.Name, "name", name, "The name of the application")
	flagSet.StringVar(&flags.Name, "output", name, "Named output file. Deprecated in favour of 'name'")
	flagSet.BoolVar(&flags.Release, "release", false, "Release mode. Prepares the application for public distribution")
	flagSet.StringVar(&flags.RootDir, "dir", rootDir, "Fyne app root directory")
	flagSet.BoolVar(&flags.Silent, "silent", false, "Silent mode")
	flagSet.BoolVar(&flags.Debug, "debug", false, "Debug mode")
	flagSet.BoolVar(&flags.Pull, "pull", false, "Attempt to pull a newer version of the docker image")
	flagSet.StringVar(&flags.DockerRegistry, "docker-registry", "docker.io", "The docker registry to be used instead of dockerhub (only used with defualt docker images)")
	flagSet.BoolVar(&flags.NoNetwork, "no-network", false, "If set, the build will be done without network access")
	flagSet.StringVar(&flags.ExtraMount, "extra-mount", "", "If set, mount extra volumes on docker")
	return flags, nil
}

func defaultName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get the path for current directory %s", err)
	}
	_, output := filepath.Split(wd)
	return output, nil
}

// engineFlag is a custom flag used to define custom engine variables
type engineFlag struct {
	Engine
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ef *engineFlag) String() string {
	return fmt.Sprint(*ef)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
func (ef *engineFlag) Set(value string) error {
	var err error
	ef.Engine, err = MakeEngine(value)
	return err
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
func (ef *envFlag) Set(value string) error {
	if !strings.Contains(value, "=") {
		return errors.New("env var must defined as KEY=VALUE or KEY=")
	}
	*ef = append(*ef, value)

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
		*af = append(*af, strings.TrimSpace(v))
	}
	return nil
}

// tagsFlag is a custom flag used to define build tags
type tagsFlag []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (tf *tagsFlag) String() string {
	return fmt.Sprint(*tf)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (tf *tagsFlag) Set(value string) error {
	*tf = []string{}
	if len(*tf) > 1 {
		return errors.New("flag already set")
	}

	for _, v := range strings.Split(value, ",") {
		*tf = append(*tf, strings.TrimSpace(v))
	}
	return nil
}
