package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	// windowsOS it the windows OS name
	windowsOS = "windows"
)

var (
	// windowsArchSupported defines the supported target architectures on windows
	windowsArchSupported = []Architecture{ArchAmd64, Arch386}
	// windowsImage is the fyne-cross image for the Windows OS
	windowsImage = "fyneio/fyne-cross:1.3-windows"
)

// Windows build and package the fyne app for the windows OS
type Windows struct {
	CmdContext []Context
}

// Name returns the one word command name
func (cmd *Windows) Name() string {
	return "windows"
}

// Description returns the command description
func (cmd *Windows) Description() string {
	return "Build and package a fyne application for the windows OS"
}

// Parse parses the arguments and set the usage for the command
func (cmd *Windows) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &windowsFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}

	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, windowsArchSupported))
	flagSet.BoolVar(&flags.Console, "console", false, "If set writes a 'console binary' instead of 'GUI binary'")

	// flags used only in release mode
	flagSet.StringVar(&flags.Certificate, "certificate", "", "The name of the certificate to sign the build")
	flagSet.StringVar(&flags.Developer, "developer", "", "The developer identity for your Microsoft store account")
	flagSet.StringVar(&flags.Password, "password", "", "The password for the certificate used to sign the build")

	// Add exe extension to default output
	flagName := flagSet.Lookup("name")
	flagName.DefValue = fmt.Sprintf("%s.exe", flagName.DefValue)
	flagName.Value.Set(flagName.DefValue)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := makeWindowsContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.CmdContext = ctx
	return nil
}

// Run runs the command
func (cmd *Windows) Run() error {

	for _, ctx := range cmd.CmdContext {

		err := bumpFyneAppBuild(ctx)
		if err != nil {
			log.Infof("[i] FyneApp.toml: unable to bump the build number. Error: %s", err)
		}

		log.Infof("[i] Target: %s/%s", ctx.OS, ctx.Architecture)
		log.Debugf("%#v", ctx)

		//
		// pull image, if requested
		//
		err = pullImage(ctx)
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

		// Release mode
		if ctx.Release {
			if runtime.GOOS != windowsOS {
				return fmt.Errorf("windows release build is supported only on windows hosts")
			}

			err = fyneReleaseHost(ctx)
			if err != nil {
				return fmt.Errorf("could not package the Fyne app: %v", err)
			}

			packageName := ctx.Name + ".appx"
			if pos := strings.LastIndex(ctx.Name, ".exe"); pos > 0 {
				packageName = ctx.Name[:pos] + ".appx"
			}

			// move the dist package into the "dist" folder
			srcFile := volume.JoinPathHost(ctx.WorkDirHost(), packageName)
			distFile := volume.JoinPathHost(ctx.DistDirHost(), ctx.ID, packageName)
			err = os.MkdirAll(filepath.Dir(distFile), 0755)
			if err != nil {
				return fmt.Errorf("could not create the dist package dir: %v", err)
			}

			err = os.Rename(srcFile, distFile)
			if err != nil {
				return err
			}
			return nil
		}

		// Build mode
		windres, err := WindowsResource(ctx)
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

		// remove the windres file under the project root
		os.Remove(volume.JoinPathHost(ctx.WorkDirHost(), windres))

		// create a zip archive from the compiled binary under the "bin" folder
		// and place it under the "dist" folder
		srcFile := volume.JoinPathHost(ctx.BinDirHost(), ctx.ID, ctx.Name)
		distFile := volume.JoinPathHost(ctx.DistDirHost(), ctx.ID, ctx.Name+".zip")
		err = os.MkdirAll(filepath.Dir(distFile), 0755)
		if err != nil {
			return fmt.Errorf("could not create the dist package dir: %v", err)
		}
		err = volume.Zip(srcFile, distFile)
		if err != nil {
			return err
		}

		log.Infof("[✓] Package: %s", distFile)
	}

	return nil
}

// Usage displays the command usage
func (cmd *Windows) Usage() {
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

// windowsFlags defines the command-line flags for the windows command
type windowsFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag

	// Console defines if the Windows app will build as "console binary" instead of "GUI binary"
	Console bool

	//Certificate represents the name of the certificate to sign the build
	Certificate string
	//Developer represents the developer identity for your Microsoft store account
	Developer string
	//Password represents the password for the certificate used to sign the build [Windows]
	Password string
}

// makeWindowsContext returns the command context for a windows target
func makeWindowsContext(flags *windowsFlags, args []string) ([]Context, error) {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, windowsArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make build context for %s OS: %s", windowsOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags, args)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = windowsOS
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		ctx.Certificate = flags.Certificate
		ctx.Developer = flags.Developer
		ctx.Password = flags.Password
		ctx.Env["GOOS"] = "windows"
		switch arch {
		case ArchAmd64:
			ctx.Env["GOARCH"] = "amd64"
			ctx.Env["CC"] = "x86_64-w64-mingw32-gcc"
		case Arch386:
			ctx.Env["GOARCH"] = "386"
			ctx.Env["CC"] = "i686-w64-mingw32-gcc"
		}

		if !flags.Console {
			ctx.LdFlags = append(ctx.LdFlags, "-H=windowsgui")
		}

		// set docker registry for default images
		if flags.DockerRegistry != "" {
			windowsImage = fmt.Sprintf("%s/%s", flags.DockerRegistry, windowsImage)
		}

		if flags.DockerImage == "" {
			ctx.DockerImage = windowsImage
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
