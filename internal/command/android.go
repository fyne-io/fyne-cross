package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// androidOS is the android OS name
	androidOS = "android"
	// androidImage is the fyne-cross image for the Android OS
	androidImage = "lucor/fyne-cross:android-latest"
)

// Android build and package the fyne app for the android OS
type Android struct {
	Context
}

// Name returns the one word command name
func (cmd *Android) Name() string {
	return "android"
}

// Description returns the command description
func (cmd *Android) Description() string {
	return "Build and package a fyne application for the android OS"
}

// Parse parses the arguments and set the usage for the command
func (cmd *Android) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &androidFlags{
		CommonFlags: commonFlags,
	}

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := makeAndroidContext(flags, flagSet.Args())
	if err != nil {
		return err
	}

	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *Android) Run() error {

	ctx := cmd.Context

	log.Infof("[i] Target: %s", ctx.OS)
	log.Debugf("%#v", ctx)

	//
	// prepare build
	//
	err := cleanTargetDirs(ctx)
	if err != nil {
		return err
	}

	err = goModInit(ctx)
	if err != nil {
		return err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.apk", cmd.Context.Output)

	err = prepareIcon(ctx)
	if err != nil {
		return err
	}

	err = fynePackage(ctx)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(ctx.WorkDirHost(), ctx.Package, packageName)
	distFile := volume.JoinPathHost(ctx.DistDirHost(), ctx.ID, packageName)
	err = os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("could not create the dist package dir: %v", err)
	}

	err = os.Rename(srcFile, distFile)
	if err != nil {
		return err
	}

	log.Infof("[✓] Package: %s", distFile)
	return nil
}

// Usage displays the command usage
func (cmd *Android) Usage() {
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

Options:
`

	printUsage(template, data)
	flagSet.PrintDefaults()
}

// androidFlags defines the command-line flags for the android command
type androidFlags struct {
	*CommonFlags
}

// makeAndroidContext returns the command context for an android target
func makeAndroidContext(flags *androidFlags, args []string) (Context, error) {
	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return Context{}, err
	}

	// appID is mandatory for android
	if ctx.AppID == "" {
		return Context{}, fmt.Errorf("appID is mandatory for %s", androidOS)
	}

	ctx.OS = androidOS
	ctx.ID = androidOS

	// set context based on command-line flags
	if flags.DockerImage == "" {
		ctx.DockerImage = androidImage
	}

	return ctx, nil
}
