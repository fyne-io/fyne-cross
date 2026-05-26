package command

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/mod/semver"
	"golang.org/x/sys/execabs"
)

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
			err = os.WriteFile(volume.JoinPathHost(ctx.WorkDirHost(), ctx.Icon), icon.FyneLogo, 0o644)
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

// checkFyneBinHost checks if the fyne cli tool is installed on the host
func checkFyneBinHost(ctx Context) (string, error) {
	fyne, err := execabs.LookPath("fyne")
	if err != nil {
		return "", fmt.Errorf("missed requirement: fyne. To install: `go install fyne.io/fyne/v2/cmd/fyne@latest` and add $GOPATH/bin to $PATH")
	}

	return fyne, nil
}

func fyneCommandVersion(fyne string, ctx Context, image containerImage) string {
	var out string
	if image.OS() == "darwin" {
		buf := &bytes.Buffer{}
		cmd := exec.Command(fyne, "version")
		cmd.Stdout = buf
		if err := cmd.Run(); err != nil {
			return ""
		}
		out = buf.String()
	} else {
		out, _ = image.Command(ctx.Volume, options{}, []string{fyne, "version"})
	}
	for _, line := range strings.Split(out, "\n") {
		_, ver, found := strings.Cut(line, "fyne cli version: ")
		if found {
			return strings.TrimSuffix(ver, "\r")
		}
	}
	return ""
}

func fyneCommandVersionCompare(fyne, ver string, ctx Context, image containerImage) int {
	return semver.Compare(fyneCommandVersion(fyne, ctx, image), ver)
}

func fyneCommand(binary, command, icon string, ctx Context, image containerImage) []string {
	target := image.Target()

	appBuildOpt := "-app-build"
	appVersionOpt := "-app-version"

	if fyneCommandVersionCompare(binary, "v2.0.0", ctx, image) >= 0 {
		appBuildOpt = "-appBuild"
		appVersionOpt = "-appVersion"
	}

	args := []string{
		binary, command,
		"-os", target,
		"-name", ctx.Name,
		"-icon", icon,
		appBuildOpt, ctx.AppBuild,
		appVersionOpt, ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-id", ctx.AppID)
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
