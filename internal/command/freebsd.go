package command

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// freebsdOS it the freebsd OS name
	freebsdOS = "freebsd"
	// freebsdImageAmd64 is the fyne-cross image for the FreeBSD OS amd64 arch
	freebsdImageAmd64 = "fyneio/fyne-cross-images:freebsd-amd64"
	// freebsdImageArm64 is the fyne-cross image for the FreeBSD OS arm64 arch
	freebsdImageArm64 = "fyneio/fyne-cross-images:freebsd-arm64"
)

// freebsdArchSupported defines the supported target architectures on freebsd
var freebsdArchSupported = []Architecture{ArchAmd64, ArchArm64}

// FreeBSD build and package the fyne app for the freebsd OS
type freeBSD struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*freeBSD)(nil)

func FreeBSD() *cli.Command {
	cmd := &freeBSD{}

	commonFlags, cliFlags, err := newCommonFlags()
	if err != nil {
		return nil
	}

	flags := &freebsdFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}

	return &cli.Command{
		Name:  "freebsd",
		Usage: "Builds and packages a fyne application for the freebsd OS",
		Flags: append(cliFlags,
			&cli.GenericFlag{
				Destination: flags.TargetArch,
				Name:        "arch",
				Usage:       fmt.Sprintf(`set list of target architecture to build separated by comma; supported: %s`, freebsdArchSupported),
			},
		),
		Action: func(ctx *cli.Context) error {
			if err := cmd.setupContainerImages(flags, ctx.Args().Slice()); err != nil {
				return err
			}
			return cmd.run()
		},
	}
}

func (cmd *freeBSD) run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Run runs the command
func (cmd *freeBSD) Build(image containerImage) (string, error) {
	//
	// package
	//
	log.Info("[i] Packaging app...")
	packageName := fmt.Sprintf("%s.tar.xz", cmd.defaultContext.Name)

	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	if cmd.defaultContext.Release {
		// Release mode
		err = fyneRelease(cmd.defaultContext, image)
	} else {
		// Build mode
		err = fynePackage(cmd.defaultContext, image)
	}
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}
	image.Run(cmd.defaultContext.Volume, options{}, []string{
		"mv",
		volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), packageName),
		volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
	})

	// Extract the resulting executable from the tarball
	image.Run(cmd.defaultContext.Volume,
		options{WorkDir: volume.JoinPathContainer(cmd.defaultContext.BinDirContainer(), image.ID())},
		[]string{
			"tar", "-xf",
			volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
			"--strip-components=3", "usr/local/bin",
		})

	return packageName, nil
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
			image.SetEnv("CC", "clang --sysroot=/freebsd --target=x86_64-unknown-freebsd12")
			image.SetEnv("CXX", "clang++ --sysroot=/freebsd --target=x86_64-unknown-freebsd12")
			if runtime.GOARCH == string(ArchArm64) {
				if v, ok := ctx.Env["CGO_LDFLAGS"]; ok {
					image.SetEnv("CGO_LDFLAGS", v+" -fuse-ld=lld")
				} else {
					image.SetEnv("CGO_LDFLAGS", "-fuse-ld=lld")
				}
			}
		case ArchArm64:
			image = runner.createContainerImage(arch, freebsdOS, overrideDockerImage(flags.CommonFlags, freebsdImageArm64))
			image.SetEnv("GOARCH", "arm64")
			if v, ok := ctx.Env["CGO_LDFLAGS"]; ok {
				image.SetEnv("CGO_LDFLAGS", v+" -fuse-ld=lld")
			} else {
				image.SetEnv("CGO_LDFLAGS", "-fuse-ld=lld")
			}
			image.SetEnv("CC", "clang --sysroot=/freebsd --target=aarch64-unknown-freebsd12")
			image.SetEnv("CXX", "clang++ --sysroot=/freebsd --target=aarch64-unknown-freebsd12")
		}
		image.SetEnv("GOOS", "freebsd")

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}
