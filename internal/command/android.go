package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// androidOS is the android OS name
	androidOS = "android"
	// androidImage is the fyne-cross image for the Android OS
	androidImage = baseImage + "-android"
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

	cmdCtx, err := makeAndroidContext(flags)
	if err != nil {
		return err
	}

	cmd.Context = cmdCtx
	return nil
}

// Run runs the command
func (cmd *Android) Run() error {

	cmdCtx := cmd.Context

	log.Infof("[i] Target: %s", cmdCtx.OS)
	log.Debugf("%#v", cmdCtx)

	//
	// prepare build
	//
	err := cmdCtx.CleanTempTargetDir()
	if err != nil {
		return err
	}

	err = GoModInit(cmdCtx)
	if err != nil {
		return err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.apk", cmd.Context.Output)

	// copy the icon to tmp dir
	err = volume.Copy(cmd.Context.Icon, volume.JoinPathHost(cmdCtx.TmpDirHost(), cmdCtx.ID, icon.Default))
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app due to error copying the icon: %v", err)
	}

	err = FynePackage(cmdCtx)
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(cmdCtx.WorkDirHost(), packageName)
	distFile := volume.JoinPathHost(cmdCtx.DistDirHost(), cmdCtx.ID, packageName)
	err = os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("Could not create the dist package dir: %v", err)
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
Usage: fyne-cross {{ .Name }} [options] 

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
func makeAndroidContext(flags *androidFlags) (Context, error) {
	ctx, err := makeDefaultContext(flags.CommonFlags)
	if err != nil {
		return Context{}, err
	}

	// appID is mandatory for android
	if ctx.AppID == "" {
		return Context{}, fmt.Errorf("appID is mandatory for %s", androidOS)
	}

	ctx.DockerImage = androidImage
	ctx.ID = androidOS
	ctx.OS = androidOS
	return ctx, nil
}
