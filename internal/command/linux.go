package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// linuxOS it the linux OS name
	linuxOS = "linux"
	// linuxImage is the fyne-cross image for the Linux OS
	linuxImageAmd64 = "lucor/fyne-cross:base-latest"
	linuxImage386   = "lucor/fyne-cross:linux-386-latest"
	linuxImageArm64 = "lucor/fyne-cross:linux-arm64-latest"
	linuxImageArm   = "lucor/fyne-cross:linux-arm-latest"
)

var (
	// linuxArchSupported defines the supported target architectures on linux
	linuxArchSupported = []Architecture{ArchAmd64, Arch386, ArchArm, ArchArm64}
)

// Linux build and package the fyne app for the linux OS
type Linux struct {
	Context []Context
}

// Name returns the one word command name
func (cmd *Linux) Name() string {
	return "linux"
}

// Description returns the command description
func (cmd *Linux) Description() string {
	return "Build and package a fyne application for the linux OS"
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
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, windowsArchSupported))

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := linuxContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *Linux) Run() error {

	for _, ctx := range cmd.Context {

		log.Infof("[i] Target: %s/%s", ctx.OS, ctx.Architecture)
		log.Debugf("%#v", ctx)

		//
		// prepare build
		//
		err := cleanTargetDirs(ctx)
		if err != nil {
			return err
		}

		err = goModInit(ctx)
		if err != nil {
			return err
		}

		//
		// build
		//
		err = goBuild(ctx)
		if err != nil {
			return err
		}

		//
		// package
		//
		log.Info("[i] Packaging app...")

		packageName := fmt.Sprintf("%s.tar.gz", ctx.Output)

		err = prepareIcon(ctx)
		if err != nil {
			return err
		}

		err = fynePackage(ctx)
		if err != nil {
			return fmt.Errorf("could not package the Fyne app: %v", err)
		}

		// move the dist package into the "dist" folder
		srcFile := volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, packageName)
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
	}

	return nil
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

// linuxContext returns the command context for a linux target
func linuxContext(flags *linuxFlags, args []string) ([]Context, error) {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, linuxArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make build context for %s OS: %s", linuxOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags, args)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = linuxOS
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		var defaultDockerImage string
		switch arch {
		case ArchAmd64:
			defaultDockerImage = linuxImageAmd64
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=amd64", "CC=gcc")
		case Arch386:
			defaultDockerImage = linuxImage386
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=386", "CC=i686-linux-gnu-gcc")
		case ArchArm:
			defaultDockerImage = linuxImageArm
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=arm", "CC=arm-linux-gnueabihf-gcc", "GOARM=7")
			ctx.Tags = []string{"gles"}
		case ArchArm64:
			defaultDockerImage = linuxImageArm64
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=arm64", "CC=aarch64-linux-gnu-gcc")
			ctx.Tags = []string{"gles"}
		}

		// set context based on command-line flags
		if flags.DockerImage == "" {
			ctx.DockerImage = defaultDockerImage
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
