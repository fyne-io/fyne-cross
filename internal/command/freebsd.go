package command

import (
	"fmt"
	"runtime"

	"github.com/fyne-io/fyne-cross/internal/log"
)

const (
	// freebsdOS it the freebsd OS name
	freebsdOS = "freebsd"
	// freebsdImageAmd64 is the fyne-cross image for the FreeBSD OS amd64 arch
	freebsdImageAmd64 = "docker.io/fyneio/fyne-cross:1.3-freebsd-amd64"
	// freebsdImageArm64 is the fyne-cross image for the FreeBSD OS arm64 arch
	freebsdImageArm64 = "docker.io/fyneio/fyne-cross:1.3-freebsd-arm64"
)

var (
	// freebsdArchSupported defines the supported target architectures on freebsd
	freebsdArchSupported = []Architecture{ArchAmd64, ArchArm64}
)

// FreeBSD build and package the fyne app for the freebsd OS
type freeBSD struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*freeBSD)(nil)
var _ Command = (*freeBSD)(nil)

func NewFreeBSD() *freeBSD {
	return &freeBSD{}
}

func (cmd *freeBSD) Name() string {
	return "freebsd"
}

// Description returns the command description
func (cmd *freeBSD) Description() string {
	return "Build and package a fyne application for the freebsd OS"
}

func (cmd *freeBSD) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *freeBSD) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &freebsdFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, freebsdArchSupported))

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *freeBSD) Build(image containerImage) (string, error) {
	//
	// build
	//
	err := goBuild(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	//
	// package
	//
	log.Info("[i] Packaging app...")

	packageName := fmt.Sprintf("%s.tar.xz", cmd.defaultContext.Name)

	err = prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	err = fynePackage(cmd.defaultContext, image)
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	return packageName, nil
}

// Usage displays the command usage
func (cmd *freeBSD) Usage() {
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

// freebsdFlags defines the command-line flags for the freebsd command
type freebsdFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag
}

// setupContainerImages returns the command context for a freebsd target
func (cmd *freeBSD) setupContainerImages(flags *freebsdFlags, args []string) error {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, freebsdArchSupported)
	if err != nil {
		return fmt.Errorf("could not make build context for %s OS: %s", freebsdOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	for _, arch := range targetArch {
		var image containerImage

		switch arch {
		case ArchAmd64:
			image = runner.createContainerImage(arch, freebsdOS, overrideDockerImage(flags.CommonFlags, freebsdImageAmd64))
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "x86_64-unknown-freebsd12-clang")
		case ArchArm64:
			image = runner.createContainerImage(arch, freebsdOS, overrideDockerImage(flags.CommonFlags, freebsdImageArm64))
			image.SetEnv("GOARCH", "arm64")
			if v, ok := ctx.Env["CGO_LDFLAGS"]; ok {
				image.SetEnv("CGO_LDFLAGS", v+" -fuse-ld=lld")
			} else {
				image.SetEnv("CGO_LDFLAGS", "-fuse-ld=lld")
			}
			image.SetEnv("CC", "aarch64-unknown-freebsd12-clang")
		}
		image.SetEnv("GOOS", "freebsd")

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
