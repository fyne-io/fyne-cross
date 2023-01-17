package command

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// darwinOS it the darwin OS name
	darwinOS = "darwin"
)

var (
	// darwinArchSupported defines the supported target architectures on darwin
	darwinArchSupported = []Architecture{ArchAmd64, ArchArm64}
	// darwinImage is the fyne-cross image for the Darwin OS
	darwinImage = "fyneio/fyne-cross:1.3-darwin"
)

// Darwin build and package the fyne app for the darwin OS
type darwin struct {
	Images         []containerImage
	defaultContext Context

	localBuild bool
}

var _ platformBuilder = (*darwin)(nil)
var _ Command = (*darwin)(nil)

func NewDarwinCommand() *darwin {
	return &darwin{localBuild: false}
}

func (cmd *darwin) Name() string {
	return "darwin"
}

// Description returns the command description
func (cmd *darwin) Description() string {
	return "Build and package a fyne application for the darwin OS"
}

func (cmd *darwin) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *darwin) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &darwinFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, darwinArchSupported))

	// Add flags to use only on darwin host
	if runtime.GOOS == darwinOS {
		flagSet.BoolVar(&cmd.localBuild, "local", true, "If set uses the fyne CLI tool installed on the host in place of the docker images")
	}

	// flags used only in release mode
	flagSet.StringVar(&flags.Category, "category", "", "The category of the app for store listing")

	flagSet.StringVar(&flags.MacOSXVersionMin, "macosx-version-min", "unset", "Specify the minimum version that the SDK you used to create the Darwin image support")

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *darwin) Build(image containerImage) (string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	var packageName string
	if cmd.defaultContext.Release {
		if runtime.GOOS != darwinOS {
			return "", fmt.Errorf("darwin release build is supported only on darwin hosts")
		}

		packageName = fmt.Sprintf("%s.pkg", cmd.defaultContext.Name)

		err = fyneReleaseHost(cmd.defaultContext, image)
		if err != nil {
			return "", fmt.Errorf("could not package the Fyne app: %v", err)
		}

	} else if cmd.localBuild {
		packageName = fmt.Sprintf("%s.app", cmd.defaultContext.Name)

		err = fynePackageHost(cmd.defaultContext, image)
		if err != nil {
			return "", fmt.Errorf("could not package the Fyne app: %v", err)
		}
	} else {
		packageName = fmt.Sprintf("%s.app", cmd.defaultContext.Name)

		err = fynePackage(cmd.defaultContext, image)
		if err != nil {
			return "", fmt.Errorf("could not package the Fyne app: %v", err)
		}
	}

	// move the dist package into the expected tmp/$ID/packageName location in the container
	image.Run(cmd.defaultContext.Volume, options{}, []string{
		"mv",
		volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), packageName),
		volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
	})

	// copy the binary into the expected bin/$ID/packageName location in the container
	image.Run(cmd.defaultContext.Volume, options{},
		[]string{
			"sh", "-c", fmt.Sprintf("cp %q %q",
				volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName, "Contents", "MacOS", "*"),
				volume.JoinPathContainer(cmd.defaultContext.BinDirContainer(), image.ID()),
			),
		})

	return packageName, nil
}

// Usage displays the command usage
func (cmd *darwin) Usage() {
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

// darwinFlags defines the command-line flags for the darwin command
type darwinFlags struct {
	*CommonFlags

	//Category represents the category of the app for store listing
	Category string

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag

	// Specify MacOSX minimum version
	MacOSXVersionMin string
}

// setupContainerImages returns the command context for a darwin target
func (cmd *darwin) setupContainerImages(flags *darwinFlags, args []string) error {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, darwinArchSupported)
	if err != nil {
		return fmt.Errorf("could not make command context for %s OS: %s", darwinOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	if ctx.AppID == "" {
		return errors.New("appID is mandatory")
	}

	ctx.Category = flags.Category

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	for _, arch := range targetArch {
		var image containerImage

		switch arch {
		case ArchAmd64:
			image = runner.createContainerImage(arch, darwinOS, overrideDockerImage(flags.CommonFlags, darwinImage))
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "o64-clang")

			if flags.MacOSXVersionMin == "unset" {
				flags.MacOSXVersionMin = "10.12"
			}
			if flags.MacOSXVersionMin != "" {
				image.SetEnv("CGO_CFLAGS", fmt.Sprintf("-mmacosx-version-min=%s", flags.MacOSXVersionMin))
				image.SetEnv("CGO_LDFLAGS", fmt.Sprintf("-mmacosx-version-min=%s", flags.MacOSXVersionMin))
			}
		case ArchArm64:
			image = runner.createContainerImage(arch, darwinOS, overrideDockerImage(flags.CommonFlags, darwinImage))
			image.SetEnv("GOARCH", "arm64")
			image.SetEnv("CC", "oa64-clang")

			if flags.MacOSXVersionMin == "unset" {
				flags.MacOSXVersionMin = "11.1"
			}
			if flags.MacOSXVersionMin != "" {
				image.SetEnv("CGO_CFLAGS", fmt.Sprintf("-mmacosx-version-min=%s", flags.MacOSXVersionMin))
				image.SetEnv("CGO_LDFLAGS", fmt.Sprintf("-fuse-ld=lld -mmacosx-version-min=%s", flags.MacOSXVersionMin))
			}
		}
		image.SetEnv("GOOS", "darwin")

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
