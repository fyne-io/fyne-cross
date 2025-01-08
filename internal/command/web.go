package command

import (
	"fmt"

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

var (
	_ platformBuilder = (*web)(nil)
	_ Command         = (*web)(nil)
)

func NewWebCommand() *web {
	return &web{}
}

// Name returns the one word command name
func (cmd *web) Name() string {
	return "web"
}

// Description returns the command description
func (cmd *web) Description() string {
	return "Build and package a fyne application for the web"
}

func (cmd *web) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *web) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &webFlags{
		CommonFlags: commonFlags,
	}

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	return cmd.setupContainerImages(flags, flagSet.Args())
}

// Run runs the command
func (cmd *web) Build(image containerImage) (string, error) {
	log.Info("[i] Packaging app...")

	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

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

// Usage displays the command usage
func (cmd *web) Usage() {
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
