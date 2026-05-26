package command

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// iosOS it the ios OS name
	iosOS = "ios"
	// iosImage is the fyne-cross image for the iOS OS
	iosImage = "fyneio/fyne-cross:1.3-base"
)

// IOS build and package the fyne app for the ios OS
type iOS struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*iOS)(nil)

func IOS() *cli.Command {
	cmd := &iOS{}

	commonFlags, cliFlags, err := newCommonFlags()
	if err != nil {
		return nil
	}

	flags := &iosFlags{
		CommonFlags: commonFlags,
	}

	return &cli.Command{
		Name:  "ios",
		Usage: "Builds and packages a fyne application for the ios OS",
		Flags: append(cliFlags,
			&cli.StringFlag{
				Name:        "certificate",
				Usage:       "set the name of the certificate to sign the build",
				Destination: &flags.Certificate,
			},
			&cli.StringFlag{
				Name:        "profile",
				Usage:       "set the name of the provisioning profile for this release build",
				Destination: &flags.Profile,
			},
		),
		Action: func(ctx *cli.Context) error {
			if err := cmd.setupContainerImages(flags, ctx.Args().Slice()); err != nil {
				return err
			}
			return cmd.run()
		},
	}
}

func (cmd *iOS) run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
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
		packageName, err = fyneReleaseHost(cmd.defaultContext, image)
	} else {
		// Build mode
		packageName, err = fynePackageHost(cmd.defaultContext, image)
	}

	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the expected tmp/$ID/packageName location in the container
	image.Run(cmd.defaultContext.Volume, options{}, []string{
		"sh", "-c", fmt.Sprintf("mv %q/*.ipa %q",
			cmd.defaultContext.WorkDirContainer(),
			volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName)),
	})

	return packageName, nil
}

// iosFlags defines the command-line flags for the ios command
type iosFlags struct {
	*CommonFlags

	// Certificate represents the name of the certificate to sign the build
	Certificate string

	// Profile represents the name of the provisioning profile for this release build
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

	ctx.Certificate = flags.Certificate
	ctx.Profile = flags.Profile

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	cmd.Images = append(cmd.Images, runner.createContainerImage("", iosOS, overrideDockerImage(flags.CommonFlags, iosImage)))

	return nil
}
