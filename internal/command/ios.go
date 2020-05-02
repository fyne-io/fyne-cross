package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// iosOS it the ios OS name
	iosOS = "ios"
	// iosImage is the fyne-cross image for the iOS OS
	iosImage = baseImage
)

// IOS build and package the fyne app for the ios OS
type IOS struct {
	Context Context
}

// Name returns the one word command name
func (cmd *IOS) Name() string {
	return "ios"
}

// Description returns the command description
func (cmd *IOS) Description() string {
	return "Build and package a fyne application for the iOS OS"
}

// Parse parses the arguments and set the usage for the command
func (cmd *IOS) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &iosFlags{
		CommonFlags: commonFlags,
	}

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s. Must match a valid provisioning profile [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	cmdCtx, err := makeIOSContext(flags)
	if err != nil {
		return err
	}
	cmd.Context = cmdCtx
	return nil
}

// Run runs the command
func (cmd *IOS) Run() error {

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

	packageName := fmt.Sprintf("%s.app", cmdCtx.Output)

	// copy the icon to tmp dir
	err = volume.Copy(cmdCtx.Icon, volume.JoinPathHost(cmdCtx.TmpDirHost(), cmdCtx.ID, icon.Default))
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
func (cmd *IOS) Usage() {
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

Note: available only on darwin hosts

Options:
`

	printUsage(template, data)
	flagSet.PrintDefaults()
}

// iosFlags defines the command-line flags for the ios command
type iosFlags struct {
	*CommonFlags
}

// makeIOSContext returns the command context for an iOS target
func makeIOSContext(flags *iosFlags) (Context, error) {

	if runtime.GOOS != darwinOS {
		return Context{}, fmt.Errorf("iOS build is supported only on darwin hosts")
	}

	ctx, err := makeDefaultContext(flags.CommonFlags)
	if err != nil {
		return Context{}, err
	}

	// appID is mandatory for ios
	if ctx.AppID == "" {
		return Context{}, fmt.Errorf("appID is mandatory for %s", iosImage)
	}

	ctx.DockerImage = iosOS
	ctx.ID = iosImage
	ctx.OS = iosImage

	return ctx, nil
}
