package command

import (
	"fmt"
	"runtime"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// iosOS it the ios OS name
	iosOS = "ios"
	// iosImage is the fyne-cross image for the iOS OS
	iosImage = "docker.io/fyneio/fyne-cross:1.3-base"
)

// IOS build and package the fyne app for the ios OS
type iOS struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*iOS)(nil)
var _ Command = (*iOS)(nil)

func NewIOSCommand() *iOS {
	return &iOS{}
}

func (cmd *iOS) Name() string {
	return "ios"
}

// Description returns the command description
func (cmd *iOS) Description() string {
	return "Build and package a fyne application for the iOS OS"
}

func (cmd *iOS) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *iOS) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &iosFlags{
		CommonFlags: commonFlags,
	}

	// flags used only in release mode
	flagSet.StringVar(&flags.Certificate, "certificate", "", "The name of the certificate to sign the build")
	flagSet.StringVar(&flags.Profile, "profile", "", "The name of the provisioning profile for this release build")

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s. Must match a valid provisioning profile [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *iOS) Build(image containerImage) (string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	log.Info("[i] Packaging app...")

	var packageName string
	if cmd.defaultContext.Release {
		// Release mode
		packageName = fmt.Sprintf("%s.ipa", cmd.defaultContext.Name)
		err = fyneReleaseHost(cmd.defaultContext, image)
	} else {
		// Build mode
		packageName = fmt.Sprintf("%s.app", cmd.defaultContext.Name)
		err = fynePackageHost(cmd.defaultContext, image)
	}

	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the expected tmp/$ID/packageName location in the container
	image.Run(cmd.defaultContext.Volume, options{}, []string{
		"mv",
		volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), packageName),
		volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
	})

	return packageName, nil
}

// Usage displays the command usage
func (cmd *iOS) Usage() {
	data := struct {
		Name        string
		Description string
	}{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}

	template := `
Usage: fyne-cross {{ .Name }} [options] [package]

{{ .Description }}

Note: available only on darwin hosts

Options:
`

	printUsage(template, data)
	flagSet.PrintDefaults()
}

// iosFlags defines the command-line flags for the ios command
type iosFlags struct {
	*CommonFlags

	//Certificate represents the name of the certificate to sign the build
	Certificate string

	//Profile represents the name of the provisioning profile for this release build
	Profile string
}

// setupContainerImages returns the command ContainerImages for an iOS target
func (cmd *iOS) setupContainerImages(flags *iosFlags, args []string) error {
	if runtime.GOOS != darwinOS {
		return fmt.Errorf("iOS build is supported only on darwin hosts")
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	// appID is mandatory for ios
	if ctx.AppID == "" {
		return fmt.Errorf("appID is mandatory for %s", iosImage)
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	cmd.Images = append(cmd.Images, runner.createContainerImage("", iosOS, overrideDockerImage(flags.CommonFlags, iosImage)))

	ctx.Certificate = flags.Certificate
	ctx.Profile = flags.Profile

	return nil
}
