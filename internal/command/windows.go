package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// windowsOS it the windows OS name
	windowsOS = "windows"
	// windowsImage is the fyne-cross image for the Windows OS
	windowsImage = baseImage
)

var (
	// windowsArchSupported defines the supported target architectures on windows
	windowsArchSupported = []Architecture{ArchAmd64, Arch386}
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

	// Add exe extension to default output
	flagOutput := flagSet.Lookup("output")
	flagOutput.DefValue = fmt.Sprintf("%s.exe", flagOutput.DefValue)
	flagOutput.Value.Set(flagOutput.DefValue)

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	cmdCtx, err := makeWindowsContext(flags)
	if err != nil {
		return err
	}
	cmd.CmdContext = cmdCtx
	return nil
}

// Run runs the command
func (cmd *Windows) Run() error {

	for _, cmdCtx := range cmd.CmdContext {

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

		outputBase := strings.TrimSuffix(cmdCtx.Output, ".exe")

		// convert the png icon to ico format and store under the temp directory
		pngIcon := cmdCtx.Icon
		icoIcon := volume.JoinPathHost(cmdCtx.TmpDirHost(), cmdCtx.ID, outputBase+".ico")
		err = icon.ConvertPngToIco(pngIcon, icoIcon)
		if err != nil {
			return fmt.Errorf("Could not create the windows ico: %v", err)
		}

		windres, err := WindowsResource(cmdCtx)

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

		// remove the windres file under the project root
		os.Remove(volume.JoinPathHost(cmdCtx.WorkDirHost(), windres))

		// create a zip archive from the compiled binary under the "bin" folder
		// and place it under the "dist" folder
		srcFile := volume.JoinPathHost(cmdCtx.BinDirHost(), cmdCtx.ID, cmdCtx.Output)
		distFile := volume.JoinPathHost(cmdCtx.DistDirHost(), cmdCtx.ID, outputBase+".zip")
		err = os.MkdirAll(filepath.Dir(distFile), 0755)
		if err != nil {
			return fmt.Errorf("Could not create the dist package dir: %v", err)
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
Usage: fyne-cross {{ .Name }} [options] 

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
}

// makeWindowsContext returns the command context for a windows target
func makeWindowsContext(flags *windowsFlags) ([]Context, error) {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, windowsArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make build context for %s OS: %s", windowsOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = windowsOS
		ctx.DockerImage = windowsImage
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		switch arch {
		case ArchAmd64:
			ctx.Env = []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"}
		case Arch386:
			ctx.Env = []string{"GOOS=windows", "GOARCH=386", "CC=i686-w64-mingw32-gcc"}
		}

		ctx.LdFlags = []string{"-H windowsgui"}

		// set context based on command-line flags
		if flags.Console {
			ctx.LdFlags = []string{}
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
