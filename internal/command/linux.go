package command

import (
	"fmt"
	"runtime"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// linuxOS it the linux OS name
	linuxOS = "linux"
	// linuxImage is the fyne-cross image for the Linux OS
	linuxImageAmd64 = "docker.io/fyneio/fyne-cross:1.2-base"
	linuxImage386   = "docker.io/fyneio/fyne-cross:1.2-linux-386"
	linuxImageArm64 = "docker.io/fyneio/fyne-cross:1.2-linux-arm64"
	linuxImageArm   = "docker.io/fyneio/fyne-cross:1.2-linux-arm"
)

var (
	// linuxArchSupported defines the supported target architectures on linux
	linuxArchSupported = []Architecture{ArchAmd64, Arch386, ArchArm, ArchArm64}
)

// Linux build and package the fyne app for the linux OS
type Linux struct {
	CrossBuilderCommand
	CrossBuilder
}

func NewLinuxCommand() *Linux {
	return &Linux{CrossBuilder: CrossBuilder{name: "linux", description: "Build and package a fyne application for the linux OS"}}
}

// Parse parses the arguments and set the usage for the command
func (cmd *Linux) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &linuxFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, linuxArchSupported))

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.makeLinuxContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command using helper code
func (cmd *Linux) Run() error {
	return cmd.RunInternal(cmd)
}

// Run runs the command
func (cmd *Linux) RunEach(image ContainerImage) (string, string, error) {
	//
	// build
	//
	err := goBuild(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.tar.xz", cmd.defaultContext.Name)

	err = prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", "", err
	}

	err = fynePackage(cmd.defaultContext, image)
	if err != nil {
		return "", "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(cmd.defaultContext.TmpDirHost(), image.GetID(), packageName)

	return srcFile, packageName, nil
}

// Usage displays the command usage
func (cmd *Linux) Usage() {
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

// linuxFlags defines the command-line flags for the linux command
type linuxFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag
}

// linuxContext returns the command ContainerImages for a linux target
func (cmd *Linux) makeLinuxContainerImages(flags *linuxFlags, args []string) error {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, linuxArchSupported)
	if err != nil {
		return fmt.Errorf("could not make build context for %s OS: %s", linuxOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	cmd.defaultContext = ctx
	runner := NewContainerEngine(ctx)

	for _, arch := range targetArch {
		var image ContainerImage

		switch arch {
		case ArchAmd64:
			image = runner.NewImageContainer(arch, linuxOS, overrideDockerImage(flags.CommonFlags, linuxImageAmd64))
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "gcc")
		case Arch386:
			image = runner.NewImageContainer(arch, linuxOS, overrideDockerImage(flags.CommonFlags, linuxImage386))
			image.SetEnv("GOARCH", "386")
			image.SetEnv("CC", "i686-linux-gnu-gcc")
		case ArchArm:
			image = runner.NewImageContainer(arch, linuxOS, overrideDockerImage(flags.CommonFlags, linuxImageArm))
			image.SetEnv("GOARCH", "arm")
			image.SetEnv("CC", "arm-linux-gnueabihf-gcc")
			image.SetEnv("GOARM", "7")
			image.AppendTag("gles")
		case ArchArm64:
			image = runner.NewImageContainer(arch, linuxOS, overrideDockerImage(flags.CommonFlags, linuxImageArm64))
			image.SetEnv("GOARCH", "arm64")
			image.SetEnv("CC", "aarch64-linux-gnu-gcc")
			image.AppendTag("gles")
		}

		image.SetEnv("GOOS", "linux")

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
