package command

import (
	"fmt"
	goimage "image"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"github.com/josephspurrier/goversioninfo"
)

const (
	// windowsOS it the windows OS name
	windowsOS = "windows"
	// windowsImage is the fyne-cross image for the Windows OS
	windowsImage = "docker.io/fyneio/fyne-cross-images:windows"
)

var (
	// windowsArchSupported defines the supported target architectures on windows
	windowsArchSupported = []Architecture{ArchAmd64, ArchArm64, Arch386}
)

// Windows build and package the fyne app for the windows OS
type windows struct {
	Images         []containerImage
	defaultContext Context
}

var _ platformBuilder = (*windows)(nil)
var _ Command = (*windows)(nil)

func NewWindowsCommand() *windows {
	return &windows{}
}

func (cmd *windows) Name() string {
	return "windows"
}

// Description returns the command description
func (cmd *windows) Description() string {
	return "Build and package a fyne application for the windows OS"
}

func (cmd *windows) Run() error {
	return commonRun(cmd.defaultContext, cmd.Images, cmd)
}

// Parse parses the arguments and set the usage for the command
func (cmd *windows) Parse(args []string) error {
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

	err = cmd.setupContainerImages(flags, flagSet.Args())
	return err
}

// Run runs the command
func (cmd *windows) Build(image containerImage) (string, error) {
	err := prepareIcon(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	// Release mode
	if cmd.defaultContext.Release {
		if runtime.GOOS != windowsOS {
			return "", fmt.Errorf("windows release build is supported only on windows hosts")
		}

		err = fyneReleaseHost(cmd.defaultContext, image)
		if err != nil {
			return "", fmt.Errorf("could not package the Fyne app: %v", err)
		}

		packageName := cmd.defaultContext.Name + ".appx"
		if pos := strings.LastIndex(cmd.defaultContext.Name, ".exe"); pos > 0 {
			packageName = cmd.defaultContext.Name[:pos] + ".appx"
		}

		// move the dist package into the expected tmp/$ID/packageName location in the container
		image.Run(cmd.defaultContext.Volume, options{}, []string{
			"mv",
			volume.JoinPathContainer(cmd.defaultContext.WorkDirContainer(), packageName),
			volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName),
		})

		return packageName, nil
	}

	// Build mode
	log.Info("[i] Creating Windows resource...")
	windres, err := makeWindowsResource(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}
	log.Infof("[✓] Windows resource created: %s", windres)

	//
	// build
	//
	err = goBuild(cmd.defaultContext, image)
	if err != nil {
		return "", err
	}

	//
	// package
	//

	log.Info("[i] Packaging app...")

	// remove the windres file under the project root
	log.Info("[i] Removing Windows resource")
	err = os.Remove(volume.JoinPathHost(cmd.defaultContext.WorkDirHost(), windres))
	if err != nil {
		log.Infof("[✗] Unable to remove Windows resource: %s", err)
	} else {
		log.Infof("[✓] Windows resource removed")
	}

	packageName := cmd.defaultContext.Name + ".zip"

	// create a zip archive from the compiled binary under the "bin" folder
	// and place it under the tmp folder
	err = image.Run(cmd.defaultContext.Volume, options{WorkDir: volume.JoinPathContainer(cmd.defaultContext.BinDirContainer(), image.ID())}, []string{
		"zip", "-r", volume.JoinPathContainer(cmd.defaultContext.TmpDirContainer(), image.ID(), packageName), cmd.defaultContext.Name,
	})
	if err != nil {
		return "", err
	}

	return packageName, nil
}

// Usage displays the command usage
func (cmd *windows) Usage() {
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

// setupContainerImages returns the command ContainerImages for a windows target
func (cmd *windows) setupContainerImages(flags *windowsFlags, args []string) error {
	targetArch, err := targetArchFromFlag(*flags.TargetArch, windowsArchSupported)
	if err != nil {
		return fmt.Errorf("could not make build context for %s OS: %s", windowsOS, err)
	}

	ctx, err := makeDefaultContext(flags.CommonFlags, args)
	if err != nil {
		return err
	}

	ctx.Certificate = flags.Certificate
	ctx.Developer = flags.Developer
	ctx.Password = flags.Password

	if !flags.Console {
		ctx.LdFlags = append(ctx.LdFlags, "-H=windowsgui")
	}

	cmd.defaultContext = ctx
	runner, err := newContainerEngine(ctx)
	if err != nil {
		return err
	}

	for _, arch := range targetArch {
		image := runner.createContainerImage(arch, windowsOS, overrideDockerImage(flags.CommonFlags, windowsImage))

		image.SetEnv("GOOS", "windows")
		switch arch {
		case ArchAmd64:
			image.SetEnv("GOARCH", "amd64")
			image.SetEnv("CC", "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype")
		case Arch386:
			image.SetEnv("GOARCH", "386")
			image.SetEnv("CC", "zig cc -target x86-windows-gnu -Wdeprecated-non-prototype")
		case ArchArm64:
			image.SetEnv("GOARCH", "arm64")
			image.SetEnv("CC", "zig cc -target aarch64-windows-gnu -Wdeprecated-non-prototype")
		}

		cmd.Images = append(cmd.Images, image)
	}

	return nil
}

// makeWindowsResource create a windows resource under the project root
// that will be automatically linked by compliler during the build
func makeWindowsResource(ctx Context, image containerImage) (string, error) {

	imgSrc, err := os.Open(ctx.Icon)
	if err != nil {
		return "", fmt.Errorf("failed to open icon source image: %w", err)
	}
	defer imgSrc.Close()
	srcImg, _, err := goimage.Decode(imgSrc)
	if err != nil {
		return "", fmt.Errorf("failed to decode icon source image: %w", err)
	}

	icoPath := volume.JoinPathHost(ctx.TmpDirHost(), "icon.ico")
	icoFile, err := os.Create(icoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open icon ico file: %w", err)
	}

	err = ico.Encode(icoFile, srcImg)
	if err != nil {
		return "", fmt.Errorf("failed to encode icon: %w", err)
	}

	err = icoFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close icon ico file: %w", err)
	}

	ver := ctx.AppVersion
	ver_parts := strings.Split(ver, ".")
	if len(ver_parts) > 3 {
		fmt.Println("invalid version", ver)
		os.Exit(1)
	}

	var major, minor, patch int
	major, err = strconv.Atoi(ver_parts[0])
	if err != nil {
		fmt.Println("invalid major version", major)
		os.Exit(1)
	}

	if len(ver_parts) >= 2 {
		minor, err = strconv.Atoi(ver_parts[1])
		if err != nil {
			fmt.Println("invalid minor version", minor)
			os.Exit(1)
		}
	}
	if len(ver_parts) == 3 {
		patch, err = strconv.Atoi(ver_parts[2])
		if err != nil {
			fmt.Println("invalid patch version", patch)
			os.Exit(1)
		}
	}

	build, err := strconv.Atoi(ctx.AppBuild)
	if err != nil {
		fmt.Println("invalid patch version", patch)
		os.Exit(1)
	}

	fv := goversioninfo.FileVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}

	vi := &goversioninfo.VersionInfo{
		IconPath: icoPath,
		FixedFileInfo: goversioninfo.FixedFileInfo{
			FileVersion:    fv,
			ProductVersion: fv,
		},
		StringFileInfo: goversioninfo.StringFileInfo{
			InternalName:     ctx.Name,
			OriginalFilename: ctx.Name,
			FileVersion:      ver + "." + ctx.AppBuild,
			ProductName:      ctx.Name,
			ProductVersion:   ver,
		},
	}

	// Fill the structures with config data.
	vi.Build()

	// Write the data to a buffer.
	vi.Walk()

	// copy the windows resource under the package root
	// it will be automatically linked by compiler during build
	windres := ctx.Name + ".syso"
	out := volume.JoinPathHost(ctx.WorkDirHost(), filepath.Join(ctx.Package, windres))
	err = vi.WriteSyso(out, image.Architecture().String())
	return windres, err
}
