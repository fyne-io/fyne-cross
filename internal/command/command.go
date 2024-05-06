package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/sys/execabs"
)

// Command wraps the methods for a fyne-cross command
type Command interface {
	Name() string              // Name returns the one word command name
	Description() string       // Description returns the command description
	Parse(args []string) error // Parse parses the cli arguments
	Usage()                    // Usage displays the command usage
	Run() error                // Run runs the command
}

type platformBuilder interface {
	Build(image containerImage) (string, error) // Called to build each possible architecture/OS combination
}

type closer interface {
	close() error
}

func commonRun(defaultContext Context, images []containerImage, builder platformBuilder) error {
	for _, image := range images {
		log.Infof("[i] Target: %s/%s", image.OS(), image.Architecture())
		log.Debugf("%#v", image)

		err := func() error {
			defer image.(closer).close()

			//
			// prepare build
			//
			if err := image.Prepare(); err != nil {
				return err
			}

			err := cleanTargetDirs(defaultContext, image)
			if err != nil {
				return err
			}

			err = goModInit(defaultContext, image)
			if err != nil {
				return err
			}

			packageName, err := builder.Build(image)
			if err != nil {
				return err
			}

			err = image.Finalize(packageName)
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil

}

// Usage prints the fyne-cross command usage
func Usage(commands []Command) {
	template := `fyne-cross is a simple tool to cross compile Fyne applications

Usage: fyne-cross <command> [arguments]

The commands are:

{{ range $k, $cmd := . }}	{{ printf "%-13s %s\n" $cmd.Name $cmd.Description }}{{ end }}
Use "fyne-cross <command> -help" for more information about a command.
`

	printUsage(template, commands)
}

// cleanTargetDirs cleans the temp dir for the target context
func cleanTargetDirs(ctx Context, image containerImage) error {

	dirs := map[string]string{
		"bin":  volume.JoinPathContainer(ctx.BinDirContainer(), image.ID()),
		"dist": volume.JoinPathContainer(ctx.DistDirContainer(), image.ID()),
		"temp": volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID()),
	}

	log.Infof("[i] Cleaning target directories...")
	for k, v := range dirs {
		err := image.Run(ctx.Volume, options{}, []string{"rm", "-rf", v})
		if err != nil {
			return fmt.Errorf("could not clean the %q dir %s: %v", k, v, err)
		}

		err = image.Run(ctx.Volume, options{}, []string{"mkdir", "-p", v})
		if err != nil {
			return fmt.Errorf("could not create the %q dir %s: %v", k, v, err)
		}

		log.Infof("[✓] %q dir cleaned: %s", k, v)
	}

	return nil
}

// prepareIcon prepares the icon for packaging
func prepareIcon(ctx Context, image containerImage) error {
	if !ctx.NoProjectUpload {
		iconPath := ctx.Icon
		if !filepath.IsAbs(ctx.Icon) {
			iconPath = volume.JoinPathHost(ctx.WorkDirHost(), ctx.Icon)
		}

		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			if ctx.Icon != icon.Default {
				return fmt.Errorf("icon not found at %q", ctx.Icon)
			}

			log.Infof("[!] Default icon not found at %q", ctx.Icon)
			err = ioutil.WriteFile(volume.JoinPathHost(ctx.WorkDirHost(), ctx.Icon), icon.FyneLogo, 0644)
			if err != nil {
				return fmt.Errorf("could not create the temporary icon: %s", err)
			}
			log.Infof("[✓] Created a placeholder icon using Fyne logo for testing purpose")
		}
	}

	err := image.Run(ctx.Volume, options{}, []string{"cp", volume.JoinPathContainer(ctx.WorkDirContainer(), ctx.Icon), volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID(), icon.Default)})
	if err != nil {
		return fmt.Errorf("could not copy the icon to temp folder: %v", err)
	}
	return nil
}

func printUsage(template string, data interface{}) {
	log.PrintTemplate(os.Stderr, template, data)
}

// checkFyneBinHost checks if the fyne cli tool is installed on the host
func checkFyneBinHost(ctx Context) (string, error) {
	fyne, err := execabs.LookPath("fyne")
	if err != nil {
		return "", fmt.Errorf("missed requirement: fyne. To install: `go install fyne.io/fyne/v2/cmd/fyne@latest` or `go get fyne.io/fyne/v2/cmd/fyne@latest` and add $GOPATH/bin to $PATH")
	}

	if debugging() {
		out, err := execabs.Command(fyne, "version").Output()
		if err != nil {
			return fyne, fmt.Errorf("could not get fyne cli %s version: %v", fyne, err)
		}
		log.Debugf("%s", out)
	}

	return fyne, nil
}

func fyneCommand(binary, command, icon string, ctx Context, image containerImage) []string {
	target := image.Target()

	args := []string{
		binary, command,
		"-os", target,
		"-name", ctx.Name,
		"-icon", icon,
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := image.Tags()
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}

	if ctx.Metadata != nil {
		for key, value := range ctx.Metadata {
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", key, value))
		}
	}

	return args
}

// fynePackageHost package the application using the fyne cli tool from the host
// Note: at the moment this is used only for the ios builds
func fynePackageHost(ctx Context, image containerImage) (string, error) {
	fyne, err := checkFyneBinHost(ctx)
	if err != nil {
		return "", err
	}

	icon := volume.JoinPathHost(ctx.TmpDirHost(), image.ID(), icon.Default)
	args := fyneCommand(fyne, "package", icon, ctx, image)

	// ios packaging require certificate and profile for running on devices
	if image.OS() == iosOS {
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Profile != "" {
			args = append(args, "-profile", ctx.Profile)
		}
	}

	workDir := ctx.WorkDirHost()
	if image.OS() == iosOS {
		workDir = volume.JoinPathHost(workDir, ctx.Package)
	} else {
		if ctx.Package != "." {
			args = append(args, "-src", ctx.Package)
		}
	}

	// when using local build, do not assume what CC is available and rely on os.Env("CC") is necessary
	image.UnsetEnv("CC")
	image.UnsetEnv("CGO_CFLAGS")
	image.UnsetEnv("CGO_LDFLAGS")

	if ctx.CacheDirHost() != "" {
		image.SetEnv("GOCACHE", ctx.CacheDirHost())
	}

	// run the command from the host
	fyneCmd := execabs.Command(args[0], args[1:]...)
	fyneCmd.Dir = workDir
	fyneCmd.Stdout = os.Stdout
	fyneCmd.Stderr = os.Stderr
	fyneCmd.Env = append(os.Environ(), image.AllEnv()...)

	if debugging() {
		log.Debug(fyneCmd)
	}

	err = fyneCmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	return searchLocalResult(volume.JoinPathHost(workDir, "*.app"))
}

// fyneReleaseHost package and release the application using the fyne cli tool from the host
// Note: at the moment this is used only for the ios and windows builds
func fyneReleaseHost(ctx Context, image containerImage) (string, error) {
	fyne, err := checkFyneBinHost(ctx)
	if err != nil {
		return "", err
	}

	icon := volume.JoinPathHost(ctx.TmpDirHost(), image.ID(), icon.Default)
	args := fyneCommand(fyne, "release", icon, ctx, image)

	workDir := ctx.WorkDirHost()

	ext := ""
	switch image.OS() {
	case darwinOS:
		if ctx.Category != "" {
			args = append(args, "-category", ctx.Category)
		}
		if ctx.Package != "." {
			args = append(args, "-src", ctx.Package)
		}
		ext = ".pkg"
	case iosOS:
		workDir = volume.JoinPathHost(workDir, ctx.Package)
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Profile != "" {
			args = append(args, "-profile", ctx.Profile)
		}
		ext = ".ipa"
	case windowsOS:
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Developer != "" {
			args = append(args, "-developer", ctx.Developer)
		}
		if ctx.Password != "" {
			args = append(args, "-password", ctx.Password)
		}
		if ctx.Package != "." {
			args = append(args, "-src", ctx.Package)
		}
		ext = ".appx"
	}

	// when using local build, do not assume what CC is available and rely on os.Env("CC") is necessary
	image.UnsetEnv("CC")
	image.UnsetEnv("CGO_CFLAGS")
	image.UnsetEnv("CGO_LDFLAGS")

	if ctx.CacheDirHost() != "" {
		image.SetEnv("GOCACHE", ctx.CacheDirHost())
	}

	// run the command from the host
	fyneCmd := execabs.Command(args[0], args[1:]...)
	fyneCmd.Dir = workDir
	fyneCmd.Stdout = os.Stdout
	fyneCmd.Stderr = os.Stderr
	fyneCmd.Env = append(os.Environ(), image.AllEnv()...)

	if debugging() {
		log.Debug(fyneCmd)
	}

	err = fyneCmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return searchLocalResult(volume.JoinPathHost(workDir, "*"+ext))
}

func searchLocalResult(path string) (string, error) {
	matches, err := filepath.Glob(path)
	if err != nil {
		return "", fmt.Errorf("could not find the file %v: %v", path, err)
	}

	// walk matches files to find the newest file
	var newest string
	var newestModTime time.Time
	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			continue
		}

		if fi.ModTime().After(newestModTime) {
			newest = match
			newestModTime = fi.ModTime()
		}
	}

	if newest == "" {
		return "", fmt.Errorf("could not find the file %v", path)
	}
	return filepath.Base(newest), nil
}
