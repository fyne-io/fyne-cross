package command

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

// Command wraps the methods for a fyne-cross command
type Command interface {
	Name() string              // Name returns the one word command name
	Description() string       // Description returns the command description
	Parse(args []string) error // Parse parses the cli arguments
	Usage()                    // Usage displays the command usage
	Run() error                // Run runs the command
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
func cleanTargetDirs(ctx Context) error {

	dirs := map[string]string{
		"bin":  volume.JoinPathHost(ctx.BinDirHost(), ctx.ID),
		"dist": volume.JoinPathHost(ctx.DistDirHost(), ctx.ID),
		"temp": volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID),
	}

	log.Infof("[i] Cleaning target directories...")
	for k, v := range dirs {
		err := os.RemoveAll(v)
		if err != nil {
			return fmt.Errorf("could not clean the %q dir %s: %v", k, v, err)
		}

		err = os.MkdirAll(v, 0755)
		if err != nil {
			return fmt.Errorf("could not create the %q dir %s: %v", k, v, err)
		}

		log.Infof("[✓] %q dir cleaned: %s", k, v)
	}

	return nil
}

// prepareIcon prepares the icon for packaging
func prepareIcon(ctx Context) error {

	if _, err := os.Stat(ctx.Icon); os.IsNotExist(err) {
		defaultIcon, err := volume.DefaultIconHost()
		if err != nil {
			return err
		}

		if ctx.Icon != defaultIcon {
			return fmt.Errorf("icon not found at %q", ctx.Icon)
		}

		log.Infof("[!] Default icon not found at %q", ctx.Icon)
		err = ioutil.WriteFile(ctx.Icon, icon.FyneLogo, 0644)
		if err != nil {
			return fmt.Errorf("could not create the temporary icon: %s", err)
		}
		log.Infof("[✓] Created a placeholder icon using Fyne logo for testing purpose")
	}

	if ctx.OS == "windows" {
		// convert the png icon to ico format and store under the temp directory
		icoIcon := volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, ctx.Output+".ico")
		err := icon.ConvertPngToIco(ctx.Icon, icoIcon)
		if err != nil {
			return fmt.Errorf("could not create the windows ico: %v", err)
		}
		return nil
	}

	err := volume.Copy(ctx.Icon, volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, icon.Default))
	if err != nil {
		return fmt.Errorf("could not copy the icon to temp folder: %v", err)
	}
	return nil
}

func printUsage(template string, data interface{}) {
	log.PrintTemplate(os.Stderr, template, data)
}
