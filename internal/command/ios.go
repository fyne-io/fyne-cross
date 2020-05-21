package command

import (
	"fmt"
	"os"
	"os/exec"
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
	iosImage = "lucor/fyne-cross:base-latest"
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

	ctx, err := makeIOSContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *IOS) Run() error {

	ctx := cmd.Context
	log.Infof("[i] Target: %s", ctx.OS)
	log.Debugf("%#v", ctx)

	//
	// check requirements
	// until a docker image for iOS will be available the packaging of iOS app
	// is supported only on darwin hosts via the fyne command.
	// Requirements:
	// - fyne binary
	// - XCode
	//
	fyne, err := exec.LookPath("fyne")
	if err != nil {
		return fmt.Errorf("missed requirement: fyne. To install: `go get fyne.io/fyne/cmd/fyne` and add $GOPATH/bin to $PATH")
	}

	err = exec.Command("xcrun", "xcodebuild", "-version").Run()
	if err != nil {
		return fmt.Errorf("missed requirement: XCode")
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

	//
	// package
	//

	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.app", ctx.Output)

	err = prepareIcon(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"package",
		"-os", ctx.OS,
		"-name", ctx.Output,
		"-icon", volume.JoinPathContainer(ctx.TmpDirHost(), ctx.ID, icon.Default),
		"-appID", ctx.AppID,
	}

	// run the command inside the container
	fynePackageCmd := exec.Command(fyne, args...)
	fynePackageCmd.Dir = ctx.WorkDirHost()
	fynePackageCmd.Stdout = os.Stdout
	fynePackageCmd.Stderr = os.Stderr

	if ctx.Debug {
		log.Debug(fynePackageCmd)
	}

	err = fynePackageCmd.Run()
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(ctx.WorkDirHost(), packageName)
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
func (cmd *IOS) Usage() {
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
}

// makeIOSContext returns the command context for an iOS target
func makeIOSContext(flags *iosFlags, args []string) (Context, error) {

	if runtime.GOOS == darwinOS {
		return Context{}, fmt.Errorf("iOS build is supported only on darwin hosts")
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return Context{}, err
	}

	// appID is mandatory for ios
	if ctx.AppID == "" {
		return Context{}, fmt.Errorf("appID is mandatory for %s", iosImage)
	}

	ctx.OS = iosOS
	ctx.ID = iosOS

	// set context based on command-line flags
	if flags.DockerImage == "" {
		ctx.DockerImage = iosImage
	}

	return ctx, nil
}
