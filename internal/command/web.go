package command

import (
	"fmt"
	"os"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// webOS it the ios OS name
	webOS = "web"
	// webImage is the fyne-cross image for the web
	webImage = "docker.io/fyneio/fyne-cross:1.3-web"
)

// Web build and package the fyne app for the web
type Web struct {
	Context Context
}

// Name returns the one word command name
func (cmd *Web) Name() string {
	return "web"
}

// Description returns the command description
func (cmd *Web) Description() string {
	return "Build and package a fyne application for the web"
}

// Parse parses the arguments and set the usage for the command
func (cmd *Web) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &webFlags{
		CommonFlags: commonFlags,
	}

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := makeWebContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *Web) Run() error {

	ctx := cmd.Context
	log.Infof("[i] Target: %s", ctx.OS)
	log.Debugf("%#v", ctx)

	//
	// pull image, if requested
	//
	err := pullImage(ctx)
	if err != nil {
		return err
	}

	//
	// prepare build
	//
	err = cleanTargetDirs(ctx)
	if err != nil {
		return err
	}

	err = goModInit(ctx)
	if err != nil {
		return err
	}

	err = prepareIcon(ctx)
	if err != nil {
		return err
	}

	log.Info("[i] Packaging app...")

	if ctx.Release {
		// Release mode
		err = fyneRelease(ctx)
	} else {
		// Build mode
		err = fynePackage(ctx)
	}

	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(ctx.WorkDirHost(), ctx.Package, "web")
	distFile := volume.JoinPathHost(ctx.DistDirHost(), ctx.ID)

	os.RemoveAll(distFile)

	err = os.Rename(srcFile, distFile)
	if err != nil {
		return err
	}

	log.Infof("[✓] Package: %s", distFile)
	return nil
}

// Usage displays the command usage
func (cmd *Web) Usage() {
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
func makeWebContext(flags *webFlags, args []string) (Context, error) {

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return Context{}, err
	}

	ctx.OS = webOS
	ctx.ID = webOS

	// set context based on command-line flags
	if flags.DockerImage == "" {
		ctx.DockerImage = webImage
	}

	return ctx, nil
}
