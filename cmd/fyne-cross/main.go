/*
fyne-cross is a simple tool to cross compile Fyne applications

It has been inspired by xgo and uses a docker image built on top of the
golang-cross image, that includes the MinGW compiler for windows, and an OSX
SDK, along the Fyne requirements.

Supported targets are:
  -  darwin/amd64
  -  darwin/386
  -  linux/amd64
  -  linux/386
  -  linux/arm
  -  linux/arm64
  -  windows/amd64
  -  windows/386

References
- fyne: https://fyne.io
- xgo: https://github.com/karalabe/xgo
- golang-cross: https://github.com/docker/golang-cross
- fyne-cross docker images: https://hub.docker.com/r/lucor/fyne-cross
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"fyne.io/fyne/theme"
	"github.com/lucor/fyne-cross/internal/builder"
	"github.com/lucor/fyne-cross/internal/volume"
)

const version = "develop"

// supportedTargets represents the list of supported GOARCH for a GOOS
var supportedTargets = map[string][]string{
	"darwin":  {"amd64", "386"},
	"linux":   {"amd64", "386", "arm", "arm64"},
	"windows": {"amd64", "386"},
	"android": {},
}

var (
	// targetList represents a list of target to build on separated by comma
	targetList string
	// output represents the named output file
	output string
	// pkg represents the package to build
	pkg string
	// rootDir represents the project root directory
	rootDir string
	// cache represents a directory used to share/cache sources and
	// dependencies. Default to system cache directory (i.e.
	// $HOME/.cache/fyne-cross).
	cacheDir string
	// verbosity represents the verbosity setting
	verbose bool
	// ldflags represents the flags to pass to the external linker
	ldflags string
	// printVersion if true will print the fyne-cross version
	printVersion bool
	// noStripDebug if true will not strip debug information from binaries
	noStripDebug bool
	// icon represents the application icon used for distribution
	icon string
	// appID represents the application ID used for distribution
	appID string
)

func main() {
	flag.Usage = printUsage

	defaultTarget := strings.Join([]string{build.Default.GOOS, build.Default.GOARCH}, "/")
	flag.StringVar(&targetList, "targets", defaultTarget, fmt.Sprintf("The list of targets to build separated by comma. Default to current GOOS/GOARCH %s", defaultTarget))
	flag.StringVar(&output, "output", "", "The named output file. Default to package name")
	flag.StringVar(&rootDir, "dir", "", "The root directory. Default current dir")
	flag.StringVar(&cacheDir, "cache", "", "Directory used to share/cache sources and dependencies. Default to system cache directory (i.e. $HOME/.cache/fyne-cross). To disable set to 'no'")
	flag.BoolVar(&verbose, "v", false, "Enable verbosity flag for go commands. Default to false")
	flag.StringVar(&ldflags, "ldflags", "", "Flags to pass to the external linker")
	flag.BoolVar(&noStripDebug, "no-strip", false, "If set will not strip debug information from binaries")
	flag.BoolVar(&printVersion, "version", false, "Print fyne-cross version")
	flag.StringVar(&icon, "icon", "Icon.png", "Application icon used for distribution. Default to Icon.png")
	flag.StringVar(&appID, "appID", "", "Application ID used for distribution. For ios or darwin targets an appID is required, for ios this must match a valid provisioning profile")

	flag.Parse()

	args := flag.Args()
	if len(args) > 1 {
		printUsage()
		os.Exit(2)
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	if args[0] == "help" {
		printUsage()
		os.Exit(2)
	}

	run(args)
}

func printUsage() {
	fmt.Println("Usage: fyne-cross [parameters] package")
	fmt.Println()
	fmt.Println("Cross compile a Fyne application")
	fmt.Println()

	fmt.Println("Package is the relative path to main.go file or main package. Default to '.'")
	fmt.Println()

	fmt.Println("Optional parameters:")
	flag.PrintDefaults()
	fmt.Println()

	fmt.Println("Supported targets:")
	for os, archs := range supportedTargets {
		if len(archs) == 0 {
			fmt.Printf(" - %s\n", os)
			continue
		}
		for _, arch := range archs {
			fmt.Printf(" - %s/%s\n", os, arch)
		}
	}
	fmt.Println()

	fmt.Println("Example: fyne-cross --targets=linux/amd64,windows/amd64 --output=test ./cmd/test")
}

func getVersion() string {
	if version != "develop" {
		return version
	}
	// dev version, try to get additional info
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return version
}

func run(args []string) {

	// Prints the version and exit
	if printVersion == true {
		fmt.Printf("fyne-cross version %s\n", getVersion())
		os.Exit(2)
	}

	// Check if all requirements are satisfied
	err := checkRequirements()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Parse and validate specified targets
	targets, err := parseTargets(targetList)
	if err != nil {
		fmt.Printf("Unable to parse targets option: %s\n", err)
		os.Exit(1)
	}

	// Prepare the fyne-cross layout
	vol, err := volume.Mount(rootDir, cacheDir)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = vol.CreateHostDirs()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// if the package is not set, use the current directory
	pkg := args[0]
	if pkg == "" {
		pkg = "."
	}

	// if the output is not set, use to current directory name
	if output == "" {
		wd := pkg
		if wd == "." {
			wd, err = os.Getwd()
			if err != nil {
				fmt.Printf("Cannot get the path for current directory %s\n", err)
				os.Exit(1)
			}
		}
		_, output = filepath.Split(wd)
	}

	// Print build summary
	fmt.Println("Project root dir:", vol.WorkDirHost())
	fmt.Println("Package dir:", pkg)
	fmt.Println("Bin output folder:", vol.BinDirHost())
	fmt.Println("Dist output folder: ", vol.DistDirHost())
	if icon == "Icon.png" {
		icon = volume.JoinPathHost(vol.WorkDirHost(), "Icon.png")
		if _, err := os.Stat(icon); os.IsNotExist(err) {
			fmt.Printf("[WARN] Icon app was not specified with --icon and a default one was not found at %q. Using Fyne logo as icon app for testing purpose\n", icon)
			icon = volume.JoinPathHost(vol.TmpDirHost(), "fyne.png")
			err := ioutil.WriteFile(icon, theme.FyneLogo().Content(), 0644)
			if err != nil {
				fmt.Println("Could not create the temporary icon:", err)
				os.Exit(1)
			}
		}

	}
	fmt.Println("Icon app: ", icon)

	for _, target := range targets {
		fmt.Printf("Building for %s %s\n", target[0], target[1])

		var b builder.Builder
		switch target[0] {
		case "darwin":
			b = builder.NewDarwin(target[1], output)
		case "linux":
			b = builder.NewLinux(target[1], output)
		case "windows":
			b = builder.NewWindows(target[1], output)
		case "android":
			b = builder.NewAndroid(target[1], output)
		default:
			fmt.Println("No builder defined for OS target", target[0])
			os.Exit(1)
		}

		preBuildOpts := builder.PreBuildOptions{
			Verbose: verbose,
			Icon:    icon,
			AppID:   appID,
		}
		err := b.PreBuild(vol, preBuildOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		buildOpts := builder.BuildOptions{
			Package:    pkg,
			LdFlags:    []string{ldflags},
			StripDebug: !noStripDebug,
			Verbose:    verbose,
		}
		err = b.Build(vol, buildOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		packageOpts := builder.PackageOptions{
			Icon:    icon,
			AppID:   appID,
			Verbose: verbose,
		}
		err = b.Package(vol, packageOpts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Target %s %s [OK]\n", target[0], target[1])
	}
}

// checkRequirements checks if all the build requirements are satisfied
func checkRequirements() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("Missed requirement: docker binary not found in PATH")
	}

	var stderr bytes.Buffer
	cmd := exec.Command("docker", "version")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s", stderr.Bytes())
	}
	return nil
}

// parseTargets parse comma separated target list and validate against the supported targets
func parseTargets(targetList string) ([][2]string, error) {
	targets := [][2]string{}

Parse:
	for _, target := range strings.Split(targetList, ",") {
		target = strings.TrimSpace(target)

		if target == "android" {
			targets = append(targets, [2]string{"android", ""})
			continue
		}

		osAndArch := strings.Split(target, "/")
		if len(osAndArch) != 2 {
			return targets, fmt.Errorf("unsupported target %q", target)
		}

		targetOS, targetArch := osAndArch[0], osAndArch[1]
		supportedArchs, ok := supportedTargets[targetOS]
		if !ok {
			return targets, fmt.Errorf("unsupported os %q in target %q", targetOS, target)
		}

		if targetArch == "*" {
			for _, arch := range supportedArchs {
				targets = append(targets, [2]string{targetOS, arch})
			}
			continue
		}

		for _, arch := range supportedArchs {
			if targetArch == arch {
				targets = append(targets, [2]string{targetOS, targetArch})
				continue Parse
			}
		}

		return targets, fmt.Errorf("unsupported arch %q in target %q", targetArch, target)

	}

	return targets, nil
}
