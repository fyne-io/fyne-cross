package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	darwinImage = "fyneio/fyne-cross:darwin-latest"
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
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, darwinArchSupported))

	// flags used only in release mode
	flagSet.StringVar(&flags.Category, "category", "", "The category of the app for store listing")

	flagAppID := flagSet.Lookup("app-id")
	flagAppID.Usage = fmt.Sprintf("%s [required]", flagAppID.Usage)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := darwinContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *Darwin) Run() error {

	for _, ctx := range cmd.Context {

		log.Infof("[i] Target: %s/%s", ctx.OS, ctx.Architecture)
		log.Debugf("%#v", ctx)

		//
		// pull image, if requested
		//
		err := pullImage(ctx)
		if err != nil {
			return err
		}

		//
		// prepare build
		//
		err = cleanTargetDirs(ctx)
		if err != nil {
			return err
		}

		err = goModInit(ctx)
		if err != nil {
			return err
		}

		err = prepareIcon(ctx)
		if err != nil {
			return err
		}

		//
		// package
		//
		log.Info("[i] Packaging app...")

		var packageName string
		var srcFile string
		if ctx.Release {
			if runtime.GOOS != darwinOS {
				return fmt.Errorf("darwin release build is supported only on darwin hosts")
			}

			packageName = fmt.Sprintf("%s.pkg", ctx.Name)
			srcFile = volume.JoinPathHost(ctx.WorkDirHost(), packageName)

			err = fyneReleaseHost(ctx)
			if err != nil {
				return fmt.Errorf("could not package the Fyne app: %v", err)
			}
		} else {
			err = goBuild(ctx)
			if err != nil {
				return err
			}

			packageName = fmt.Sprintf("%s.app", ctx.Name)
			srcFile = volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, packageName)

			err = fynePackage(ctx)
			if err != nil {
				return fmt.Errorf("could not package the Fyne app: %v", err)
			}
		}

		// move the package into the "dist" folder
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
func darwinContext(flags *darwinFlags, args []string) ([]Context, error) {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, darwinArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make command context for %s OS: %s", darwinOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags, args)
		if err != nil {
			return ctxs, err
		}
		if ctx.AppID == "" {
			return ctxs, errors.New("appID is mandatory")
		}

		ctx.Architecture = arch
		ctx.OS = darwinOS
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)
		ctx.Category = flags.Category

		switch arch {
		case ArchAmd64:
			ctx.Env = append(ctx.Env, "GOOS=darwin", "CGO_CFLAGS=-mmacosx-version-min=10.12", "CGO_LDFLAGS=-mmacosx-version-min=10.12", "GOARCH=amd64", "CC=o64-clang")
		case ArchArm64:
			ctx.Env = append(ctx.Env, "GOOS=darwin", "CGO_CFLAGS=-mmacosx-version-min=10.12", "CGO_LDFLAGS=-mmacosx-version-min=10.12", "GOARCH=arm64", "CC=o64-clang")
		}

		// set context based on command-line flags
		if flags.DockerImage == "" {
			ctx.DockerImage = darwinImage
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
