package command

import (
	"fmt"
	"os"
	"path"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// androidOS is the android OS name
	androidOS = "android"
	// androidImage is the fyne-cross image for the Android OS
	androidImage = "fyneio/fyne-cross-images:android"
)

var (
	// androidArchSupported defines the supported target architectures for the android OS
	androidArchSupported = []Architecture{ArchMultiple, ArchAmd64, Arch386, ArchArm, ArchArm64}
)

// Android build and package the fyne app for the android OS
type android struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*android)(nil)
var _ Command = (*android)(nil)

func NewAndroidCommand() *android {
	return &android{}
}

func (cmd *android) Name() string {
	return "android"
}

// Description returns the command description
func (cmd *android) Description() string {
	return "Build and package a fyne application for the android OS"
}

func (cmd *android) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *android) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &androidFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{string(ArchMultiple)},
	}

	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s.`, androidArchSupported))
	flagSet.StringVar(&flags.Keystore, "keystore", "", "The location of .keystore file containing signing information")
	flagSet.StringVar(&flags.KeystorePass, "keystore-pass", "", "Password for the .keystore file")
	flagSet.StringVar(&flags.KeyPass, "key-pass", "", "Password for the signer's private key, which is needed if the private key is password-protected")
	flagSet.StringVar(&flags.KeyName, "key-name", "", "Name of the key to use for signing")

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *android) Build(image containerImage) (string, error) {
	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.apk", cmd.defaultContext.Name)
	pattern := "*.apk"

	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	if cmd.defaultContext.Release {
		err = fyneRelease(cmd.defaultContext, image)
		packageName = fmt.Sprintf("%s.aab", cmd.defaultContext.Name)
		pattern = "*.aab"
	} else {
		err = fynePackage(cmd.defaultContext, image)
	}
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	// The fyne tool sanitizes the package name to be acceptable as a
	// android package name. For details, see:
	// https://github.com/fyne-io/fyne/blob/v1.4.0/cmd/fyne/internal/mobile/build_androidapp.go#L297
	// To avoid to duplicate the fyne tool sanitize logic here, the location of
	// the dist package to move will be detected using a matching pattern
	command := fmt.Sprintf("mv %q/%s %q",
		volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), cmd.defaultContext.Package),
		pattern,
		volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
	)

	// move the dist package into the expected tmp/$ID/packageName location in the container
	// We use the shell to do the globbing and copy the file
	err = image.Run(cmd.defaultContext.Volume, options{}, []string{
		"sh", "-c", command,
	})

	if err != nil {
		return "", fmt.Errorf("could not retrieve the packaged apk")
	}

	return packageName, nil
}

// Usage displays the command usage
func (cmd *android) Usage() {
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
	KeyName      string //Name of the key to use for signing

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag
}

// setupContainerImages returns the command context for an android target
func (cmd *android) setupContainerImages(flags *androidFlags, args []string) error {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, androidArchSupported)
	if err != nil {
		return fmt.Errorf("could not make build context for %s OS: %s", androidOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	// appID is mandatory for android
	if ctx.AppID == "" {
		return fmt.Errorf("appID is mandatory for %s", androidOS)
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	for _, arch := range targetArch {
		// By default, the fyne cli tool builds a fat APK for all supported
		// instruction sets (arm, 386, amd64, arm64). A subset of instruction sets can
		// be selected by specifying target type with the architecture name.
		// E.g.: -os=android/arm
		image := runner.createContainerImage(arch, androidOS, overrideDockerImage(flags.CommonFlags, androidImage))

		if path.IsAbs(flags.Keystore) {
			return fmt.Errorf("keystore location must be relative to the project root: %s", ctx.Volume.WorkDirHost())
		}

		if !ctx.NoProjectUpload {
			if _, err := os.Stat(volume.JoinPathHost(ctx.Volume.WorkDirHost(), flags.Keystore)); err != nil {
				return fmt.Errorf("keystore location must be under the project root: %s", ctx.Volume.WorkDirHost())
			}
		}

		cmd.defaultContext.Keystore = volume.JoinPathContainer(cmd.defaultContext.Volume.WorkDirContainer(), flags.Keystore)
		cmd.defaultContext.KeystorePass = flags.KeystorePass
		cmd.defaultContext.KeyPass = flags.KeyPass
		cmd.defaultContext.KeyName = flags.KeyName

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
