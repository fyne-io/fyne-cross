package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/metadata"
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
	err := bumpFyneAppBuild(defaultContext)
	if err != nil {
		log.Infof("[i] FyneApp.toml: unable to bump the build number. Error: %s", err)
	}

	for _, image := range images {
		log.Infof("[i] Target: %s/%s", image.OS(), image.Architecture())
		log.Debugf("%#v", image)

		err = func() error {
			defer image.(closer).close()

			//
			// prepare build
			//
			if err := image.Prepare(); err != nil {
				return err
			}

			err = cleanTargetDirs(defaultContext, image)
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
		"dist": volume.JoinPathHost(ctx.DistDirContainer(), image.ID()),
		"temp": volume.JoinPathHost(ctx.TmpDirContainer(), image.ID()),
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
		if _, err := os.Stat(ctx.Icon); os.IsNotExist(err) {
			if ctx.Icon != icon.Default {
				return fmt.Errorf("icon not found at %q", ctx.Icon)
			}

			log.Infof("[!] Default icon not found at %q", ctx.Icon)
			err = ioutil.WriteFile(volume.JoinPathContainer(ctx.WorkDirHost(), ctx.Icon), icon.FyneLogo, 0644)
			if err != nil {
				return fmt.Errorf("could not create the temporary icon: %s", err)
			}
			log.Infof("[✓] Created a placeholder icon using Fyne logo for testing purpose")
		}
	}

	if image.OS() == "windows" {
		// convert the png icon to ico format and store under the temp directory
		icoIcon := volume.JoinPathHost(ctx.TmpDirHost(), image.ID(), ctx.Name+".ico")
		err := icon.ConvertPngToIco(ctx.Icon, icoIcon)
		if err != nil {
			return fmt.Errorf("could not create the windows ico: %v", err)
		}
		return nil
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
		return "", fmt.Errorf("missed requirement: fyne. To install: `go get fyne.io/fyne/v2/cmd/fyne` and add $GOPATH/bin to $PATH")
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

// fynePackageHost package the application using the fyne cli tool from the host
// Note: at the moment this is used only for the ios builds
func fynePackageHost(ctx Context, image containerImage) error {

	fyne, err := checkFyneBinHost(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"package",
		"-os", image.OS(),
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirHost(), image.ID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	// run the command from the host
	fyneCmd := execabs.Command(fyne, args...)
	fyneCmd.Dir = ctx.WorkDirHost()
	fyneCmd.Stdout = os.Stdout
	fyneCmd.Stderr = os.Stderr

	if debugging() {
		log.Debug(fyneCmd)
	}

	err = fyneCmd.Run()
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// fyneReleaseHost package and release the application using the fyne cli tool from the host
// Note: at the moment this is used only for the ios and windows builds
func fyneReleaseHost(ctx Context, image containerImage) error {

	fyne, err := checkFyneBinHost(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"release",
		"-os", image.OS(),
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirHost(), image.ID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	switch image.OS() {
	case darwinOS:
		if ctx.Category != "" {
			args = append(args, "-category", ctx.Category)
		}
	case iosOS:
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Profile != "" {
			args = append(args, "-profile", ctx.Profile)
		}
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
	}

	// run the command from the host
	fyneCmd := execabs.Command(fyne, args...)
	fyneCmd.Dir = ctx.WorkDirHost()
	fyneCmd.Stdout = os.Stdout
	fyneCmd.Stderr = os.Stderr

	if debugging() {
		log.Debug(fyneCmd)
	}

	err = fyneCmd.Run()
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// bumpFyneAppBuild increments the BuildID into the FyneApp.toml, if any,
// to behave like the fyne CLI tool
func bumpFyneAppBuild(ctx Context) error {
	data, err := metadata.LoadStandard(ctx.Volume.WorkDirHost())
	if err != nil {
		return nil // no metadata to update
	}
	data.Details.Build++
	return metadata.SaveStandard(data, ctx.Volume.WorkDirHost())
}
