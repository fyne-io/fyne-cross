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
	// darwinOS it the darwin OS name
	darwinOS = "darwin"
)

var (
	// darwinArchSupported defines the supported target architectures on darwin
	darwinArchSupported = []Architecture{ArchAmd64, Arch386}
	// darwinImage is the fyne-cross image for the Darwin OS
	darwinImage = baseImage
)

// Darwin build and package the fyne app for the darwin OS
type Darwin struct {
	Context []Context
}

// Name returns the one word command name
func (cmd *Darwin) Name() string {
	return "darwin"
}

// Description returns the command description
func (cmd *Darwin) Description() string {
	return "Build and package a fyne application for the darwin OS"
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
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, windowsArchSupported))

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	cmdCtx, err := darwinContext(flags)
	if err != nil {
		return err
	}
	cmd.Context = cmdCtx
	return nil
}

// Run runs the command
func (cmd *Darwin) Run() error {

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

		packageName := fmt.Sprintf("%s.app", cmdCtx.Output)

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
func (cmd *Darwin) Usage() {
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

// darwinFlags defines the command-line flags for the darwin command
type darwinFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag
}

// darwinContext returns the command context for a darwin target
func darwinContext(flags *darwinFlags) ([]Context, error) {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, darwinArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make command context for %s OS: %s", darwinOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = darwinOS
		ctx.DockerImage = darwinImage
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		switch arch {
		case ArchAmd64:
			ctx.Env = []string{"GOOS=darwin", "GOARCH=amd64", "CC=o32-clang"}
		case Arch386:
			ctx.Env = []string{"GOOS=darwin", "GOARCH=386", "CC=o32-clang"}
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
