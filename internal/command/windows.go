package command

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// windowsOS it the windows OS name
	windowsOS = "windows"
	// windowsImage is the fyne-cross image for the Windows OS
	windowsImage = "fyneio/fyne-cross-images:windows"
)

// windowsArchSupported defines the supported target architectures on windows
var windowsArchSupported = []Architecture{ArchAmd64, ArchArm64, Arch386}

// Windows build and package the fyne app for the windows OS
type windows struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*windows)(nil)

func Windows() *cli.Command {
	cmd := &windows{}

	commonFlags, cliFlags, err := newCommonFlags()
	if err != nil {
		return nil
	}

	flags := &windowsFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}

	return &cli.Command{
		Name:  "windows",
		Usage: "Builds and packages a fyne application for the windows OS",
		Flags: append(cliFlags,
			&cli.GenericFlag{
				Name:        "arch",
				Usage:       fmt.Sprintf("set list of target architectures to build separated by comma; supported: %s", windowsArchSupported),
				Destination: flags.TargetArch,
			},
			&cli.BoolFlag{
				Name:        "console",
				Usage:       "set to write a 'console binary' instead of 'GUI binary'",
				Destination: &flags.Console,
			},

			&cli.StringFlag{
				Name:        "certificate",
				Usage:       "set the name of the certificate to sign the build",
				Destination: &flags.Certificate,
			},
			&cli.StringFlag{
				Name:        "developer",
				Usage:       "set the developer identity for your Microsoft store account",
				Destination: &flags.Developer,
			},
			&cli.StringFlag{
				Name:        "password",
				Usage:       "set the password for the certificate used to sign the build",
				Destination: &flags.Password,
			},
		),
		Action: func(ctx *cli.Context) error {
			// XXX: flagName.DefValue = fmt.Sprintf("%s.exe", flagName.DefValue)
			if err := cmd.setupContainerImages(flags, ctx.Args().Slice()); err != nil {
				return err
			}
			return cmd.run()
		},
	}
}

func (cmd *windows) run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Run runs the command
func (cmd *windows) Build(image containerImage) (string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	// Release mode
	if cmd.defaultContext.Release {
		if runtime.GOOS != windowsOS {
			return "", fmt.Errorf("windows release build is supported only on windows hosts")
		}

		packageName, err := fyneReleaseHost(cmd.defaultContext, image)
		if err != nil {
			return "", fmt.Errorf("could not package the Fyne app: %v", err)
		}

		// move the dist package into the expected tmp/$ID/packageName location in the container
		image.Run(cmd.defaultContext.Volume, options{}, []string{
			"sh", "-c", fmt.Sprintf("mv %q/*.appx %q",
				cmd.defaultContext.WorkDirContainer(),
				volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName)),
		})

		return packageName, nil
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")
	packageName := cmd.defaultContext.Name + ".zip"

	// Build mode
	err = fynePackage(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	executableName := cmd.defaultContext.Name + ".exe"
	if pos := strings.LastIndex(cmd.defaultContext.Name, ".exe"); pos > 0 {
		executableName = cmd.defaultContext.Name
	}

	// create a zip archive from the compiled binary under the "bin" folder
	// and place it under the tmp folder
	err = image.Run(cmd.defaultContext.Volume, options{}, []string{
		"sh", "-c", fmt.Sprintf("cd %q && zip %q *.exe",
			volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), cmd.defaultContext.Package),
			volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName)),
	})
	if err != nil {
		return "", err
	}

	image.Run(cmd.defaultContext.Volume, options{}, []string{
		"sh", "-c", fmt.Sprintf("mv %q/*.exe %q",
			volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), cmd.defaultContext.Package),
			volume.JoinPathContainer(cmd.defaultContext.BinDirContainer(), image.ID(), executableName)),
	})

	return packageName, nil
}

// windowsFlags defines the command-line flags for the windows command
type windowsFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag

	// Console defines if the Windows app will build as "console binary" instead of "GUI binary"
	Console bool

	// Certificate represents the name of the certificate to sign the build
	Certificate string
	// Developer represents the developer identity for your Microsoft store account
	Developer string
	// Password represents the password for the certificate used to sign the build [Windows]
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
			image.SetEnv("CC", "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
			image.SetEnv("CXX", "zig c++ -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
		case Arch386:
			image.SetEnv("GOARCH", "386")
			image.SetEnv("CC", "zig cc -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
			image.SetEnv("CXX", "zig c++ -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
		case ArchArm64:
			image.SetEnv("GOARCH", "arm64")
			image.SetEnv("CC", "zig cc -target aarch64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
			image.SetEnv("CXX", "zig c++ -target aarch64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows")
		}

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
