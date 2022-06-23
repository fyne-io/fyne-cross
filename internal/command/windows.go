package command

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// windowsOS it the windows OS name
	windowsOS = "windows"
	// windowsImage is the fyne-cross image for the Windows OS
	windowsImage = "docker.io/fyneio/fyne-cross:1.2-windows"
)

var (
	// windowsArchSupported defines the supported target architectures on windows
	windowsArchSupported = []Architecture{ArchAmd64, Arch386}
)

// Windows build and package the fyne app for the windows OS
type windows struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*windows)(nil)
var _ Command = (*windows)(nil)

func NewWindowsCommand() *windows {
	return &windows{}
}

func (cmd *windows) Name() string {
	return "windows"
}

// Description returns the command description
func (cmd *windows) Description() string {
	return "Build and package a fyne application for the windows OS"
}

func (cmd *windows) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *windows) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &windowsFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}

	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, windowsArchSupported))
	flagSet.BoolVar(&flags.Console, "console", false, "If set writes a 'console binary' instead of 'GUI binary'")

	// flags used only in release mode
	flagSet.StringVar(&flags.Certificate, "certificate", "", "The name of the certificate to sign the build")
	flagSet.StringVar(&flags.Developer, "developer", "", "The developer identity for your Microsoft store account")
	flagSet.StringVar(&flags.Password, "password", "", "The password for the certificate used to sign the build")

	// Add exe extension to default output
	flagName := flagSet.Lookup("name")
	flagName.DefValue = fmt.Sprintf("%s.exe", flagName.DefValue)
	flagName.Value.Set(flagName.DefValue)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *windows) Build(image containerImage) (string, string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	// Release mode
	if cmd.defaultContext.Release {
		if runtime.GOOS != windowsOS {
			return "", "", fmt.Errorf("windows release build is supported only on windows hosts")
		}

		err = fyneReleaseHost(cmd.defaultContext, image)
		if err != nil {
			return "", "", fmt.Errorf("could not package the Fyne app: %v", err)
		}

		packageName := cmd.defaultContext.Name + ".appx"
		if pos := strings.LastIndex(cmd.defaultContext.Name, ".exe"); pos > 0 {
			packageName = cmd.defaultContext.Name[:pos] + ".appx"
		}

		// move the dist package into the "dist" folder
		srcFile := volume.JoinPathHost(cmd.defaultContext.WorkDirHost(), packageName)
		return srcFile, packageName, nil
	}

	// Build mode
	windres, err := WindowsResource(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	//
	// build
	//
	err = goBuild(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	//
	// package
	//

	log.Info("[i] Packaging app...")

	// remove the windres file under the project root
	os.Remove(volume.JoinPathHost(cmd.defaultContext.WorkDirHost(), windres))

	packageName := cmd.defaultContext.Name + ".zip"

	// create a zip archive from the compiled binary under the "bin" folder
	// and place it under the "dist" folder
	srcFile := volume.JoinPathHost(cmd.defaultContext.BinDirHost(), image.ID(), cmd.defaultContext.Name)
	tmpFile := volume.JoinPathHost(cmd.defaultContext.TmpDirHost(), image.ID(), packageName)
	err = volume.Zip(srcFile, tmpFile)
	if err != nil {
		return "", "", err
	}

	return tmpFile, packageName, nil
}

// Usage displays the command usage
func (cmd *windows) Usage() {
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

// windowsFlags defines the command-line flags for the windows command
type windowsFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag

	// Console defines if the Windows app will build as "console binary" instead of "GUI binary"
	Console bool

	//Certificate represents the name of the certificate to sign the build
	Certificate string
	//Developer represents the developer identity for your Microsoft store account
	Developer string
	//Password represents the password for the certificate used to sign the build [Windows]
	Password string
}

// setupContainerImages returns the command ContainerImages for a windows target
func (cmd *windows) setupContainerImages(flags *windowsFlags, args []string) error {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, windowsArchSupported)
	if err != nil {
		return fmt.Errorf("could not make build context for %s OS: %s", windowsOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	ctx.Certificate = flags.Certificate
	ctx.Developer = flags.Developer
	ctx.Password = flags.Password

	if !flags.Console {
		ctx.LdFlags = append(ctx.LdFlags, "-H=windowsgui")
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	for _, arch := range targetArch {
		image := runner.createContainerImage(arch, windowsOS, overrideDockerImage(flags.CommonFlags, windowsImage))

		image.SetEnv("GOOS", "windows")
		switch arch {
		case ArchAmd64:
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "x86_64-w64-mingw32-gcc")
		case Arch386:
			image.SetEnv("GOARCH", "386")
			image.SetEnv("CC", "i686-w64-mingw32-gcc")
		}

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
