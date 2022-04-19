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
	darwinImage = "docker.io/fyneio/fyne-cross:1.2-darwin"
)

// Darwin build and package the fyne app for the darwin OS
type Darwin struct {
	CrossBuilderCommand
	CrossBuilder

	localBuild bool
}

func NewDarwinCommand() *Darwin {
	return &Darwin{
		CrossBuilder: CrossBuilder{
			name:        "darwin",
			description: "Build and package a fyne application for the darwin OS",
		},
		localBuild: false,
	}
}

// Parse parses the arguments and set the usage for the command
func (cmd *Darwin) Parse(args []string) error {
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

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.makeDarwinContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command using helper code
func (cmd *Darwin) Run() error {
	return cmd.RunInternal(cmd)
}

// Run runs the command
func (cmd *Darwin) RunEach(image ContainerImage) (string, string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	var packageName string
	var srcFile string
	if cmd.defaultContext.Release {
		if runtime.GOOS != darwinOS {
			return "", "", fmt.Errorf("darwin release build is supported only on darwin hosts")
		}

		packageName = fmt.Sprintf("%s.pkg", cmd.defaultContext.Name)
		srcFile = volume.JoinPathHost(cmd.defaultContext.WorkDirHost(), packageName)

		err = fyneReleaseHost(cmd.defaultContext, image)
		if err != nil {
			return "", "", fmt.Errorf("could not package the Fyne app: %v", err)
		}
	} else if cmd.localBuild {
		packageName = fmt.Sprintf("%s.app", cmd.defaultContext.Name)
		srcFile = volume.JoinPathHost(cmd.defaultContext.WorkDirHost(), packageName)

		err = fynePackageHost(cmd.defaultContext, image)
		if err != nil {
			return "", "", fmt.Errorf("could not package the Fyne app: %v", err)
		}
	} else {
		err = goBuild(cmd.defaultContext, image)
		if err != nil {
			return "", "", err
		}

		packageName = fmt.Sprintf("%s.app", cmd.defaultContext.Name)
		srcFile = volume.JoinPathHost(cmd.defaultContext.TmpDirHost(), image.GetID(), packageName)

		err = fynePackage(cmd.defaultContext, image)
		if err != nil {
			return "", "", fmt.Errorf("could not package the Fyne app: %v", err)
		}
	}

	return srcFile, packageName, nil
}

// Usage displays the command usage
func (cmd *Darwin) Usage() {
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
}

// darwinContext returns the command context for a darwin target
func (cmd *Darwin) makeDarwinContainerImages(flags *darwinFlags, args []string) error {
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
	runner := NewContainerEngine(ctx)

	for _, arch := range targetArch {
		var image ContainerImage

		switch arch {
		case ArchAmd64:
			image = runner.NewContainerImage(arch, darwinOS, overrideDockerImage(flags.CommonFlags, darwinImage))
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "o64-clang")
			image.SetEnv("CGO_CFLAGS", "-mmacosx-version-min=10.12")
			image.SetEnv("CGO_LDFLAGS", "-mmacosx-version-min=10.12")
		case ArchArm64:
			image = runner.NewContainerImage(arch, darwinOS, overrideDockerImage(flags.CommonFlags, darwinImage))
			image.SetEnv("GOARCH", "arm64")
			image.SetEnv("CC", "oa64-clang")
			image.SetEnv("CGO_CFLAGS", "-mmacosx-version-min=11.1")
			image.SetEnv("CGO_LDFLAGS", "-fuse-ld=lld -mmacosx-version-min=11.1")
		}
		image.SetEnv("GOOS", "darwin")

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
