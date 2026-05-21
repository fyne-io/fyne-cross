package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// webOS it the ios OS name
	webOS = "web"
	// webImage is the fyne-cross image for the web
	webImage = "fyneio/fyne-cross-images:web"
)

// web build and package the fyne app for the web
type web struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*web)(nil)

func Web() *cli.Command {
	cmd := &web{}

	commonFlags, cliFlags, err := newCommonFlags()
	if err != nil {
		return nil
	}

	flags := &webFlags{
		CommonFlags: commonFlags,
	}

	return &cli.Command{
		Name:  "web",
		Usage: "Builds and packages a fyne application for the web",
		Flags: cliFlags,
		Action: func(ctx *cli.Context) error {
			if err := cmd.setupContainerImages(flags, ctx.Args().Slice()); err != nil {
				return err
			}
			return cmd.run()
		},
	}
}

func (cmd *web) run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Run runs the command
func (cmd *web) Build(image containerImage) (string, error) {
	log.Info("[i] Packaging app...")

	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	image.SetEnv("CGO_ENABLED", "0")
	if cmd.defaultContext.Release {
		// Release mode
		err = fyneRelease(cmd.defaultContext, image)
	} else {
		// Build mode
		err = fynePackage(cmd.defaultContext, image)
	}
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "tmp" folder
	srcFile := volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), "wasm")
	dstFile := volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID())
	return "", image.Run(cmd.defaultContext.Volume, options{}, []string{"mv", srcFile, dstFile})
}

// webFlags defines the command-line flags for the web command
type webFlags struct {
	*CommonFlags
}

// makeWebContext returns the command context for an iOS target
func (cmd *web) setupContainerImages(flags *webFlags, args []string) error {
	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	image := runner.createContainerImage("", webOS, overrideDockerImage(flags.CommonFlags, webImage))
	cmd.Images = append(cmd.Images, image)

	return nil
}
