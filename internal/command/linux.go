package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// linuxOS it the linux OS name
	linuxOS = "linux"
	// linuxImage is the fyne-cross image for the Linux OS
	linuxImage = baseImage
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

	cmdCtx, err := linuxContext(flags)
	if err != nil {
		return err
	}
	cmd.Context = cmdCtx
	return nil
}

// Run runs the command
func (cmd *Linux) Run() error {

	for _, cmdCtx := range cmd.Context {

		log.Infof("[i] Target: %s/%s", cmdCtx.OS, cmdCtx.Architecture)
		log.Debugf("%#v", cmdCtx)

		//
		// prepare build
		//
		err := cmdCtx.CleanTempTargetDir()
		if err != nil {
			return err
		}

		err = GoModInit(cmdCtx)
		if err != nil {
			return err
		}

		//
		// build
		//
		err = GoBuild(cmdCtx)
		if err != nil {
			return err
		}

		//
		// package
		//
		log.Info("[i] Packaging app...")

		packageName := fmt.Sprintf("%s.tar.gz", cmdCtx.Output)

		// copy the icon to tmp dir
		err = volume.Copy(cmdCtx.Icon, volume.JoinPathHost(cmdCtx.TmpDirHost(), cmdCtx.ID, icon.Default))
		if err != nil {
			return fmt.Errorf("Could not package the Fyne app due to error copying the icon: %v", err)
		}

		err = FynePackage(cmdCtx)
		if err != nil {
			return fmt.Errorf("Could not package the Fyne app: %v", err)
		}

		// move the dist package into the "dist" folder
		srcFile := volume.JoinPathHost(cmdCtx.TmpDirHost(), cmdCtx.ID, packageName)
		distFile := volume.JoinPathHost(cmdCtx.DistDirHost(), cmdCtx.ID, packageName)
		err = os.MkdirAll(filepath.Dir(distFile), 0755)
		if err != nil {
			return fmt.Errorf("Could not create the dist package dir: %v", err)
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
Usage: fyne-cross {{ .Name }} [options] 

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
func linuxContext(flags *linuxFlags) ([]Context, error) {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, linuxArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make build context for %s OS: %s", linuxOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = linuxOS
		ctx.DockerImage = linuxImage
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		switch arch {
		case ArchAmd64:
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=amd64", "CC=gcc")
		case Arch386:
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=386", "CC=i686-linux-gnu-gcc")
		case ArchArm:
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=arm", "CC=arm-linux-gnueabihf-gcc", "GOARM=7")
			ctx.Tags = []string{"gles"}
		case ArchArm64:
			ctx.Env = append(ctx.Env, "GOOS=linux", "GOARCH=arm64", "CC=aarch64-linux-gnu-gcc")
			ctx.Tags = []string{"gles"}
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
