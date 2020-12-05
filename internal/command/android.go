package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// androidOS is the android OS name
	androidOS = "android"
	// androidImage is the fyne-cross image for the Android OS
	androidImage = "fyneio/fyne-cross:android-latest"
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

	flagSet.StringVar(&flags.Keystore, "keystore", "", "The location of .keystore file containing signing information")
	flagSet.StringVar(&flags.KeystorePass, "keystore-pass", "", "Password for the .keystore file")
	flagSet.StringVar(&flags.KeyPass, "key-pass", "", "Password for the signer's private key, which is needed if the private key is password-protected")

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

	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.apk", cmd.Context.Name)

	err = prepareIcon(ctx)
	if err != nil {
		return err
	}

	if ctx.Release {
		err = fyneRelease(ctx)
	} else {
		err = fynePackage(ctx)
	}
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	// The fyne tool sanitizes the package name to be acceptable as a
	// android package name. For details, see:
	// https://github.com/fyne-io/fyne/blob/v1.4.0/cmd/fyne/internal/mobile/build_androidapp.go#L297
	// To avoid to duplicate the fyne tool sanitize logic here, the location of
	// the dist package to move will be detected using a matching pattern
	apkFilePattern := volume.JoinPathHost(ctx.WorkDirHost(), ctx.Package, "*.apk")
	apks, err := filepath.Glob(apkFilePattern)
	if err != nil {
		return fmt.Errorf("could not find any apk file matching %q: %v", apkFilePattern, err)
	}
	if apks == nil {
		return fmt.Errorf("could not find any apk file matching %q", apkFilePattern)
	}
	if len(apks) > 1 {
		return fmt.Errorf("multiple apk files matching %q: %v. Please remove and build again", apkFilePattern, apks)
	}
	srcFile := apks[0]
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

	Keystore     string //Keystore represents the location of .keystore file containing signing information
	KeystorePass string //Password for the .keystore file
	KeyPass      string //Password for the signer's private key, which is needed if the private key is password-protected
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

	ctx.Keystore = flags.Keystore
	ctx.KeystorePass = flags.KeystorePass
	ctx.KeyPass = flags.KeyPass

	// set context based on command-line flags
	if flags.DockerImage == "" {
		ctx.DockerImage = androidImage
	}

	return ctx, nil
}
