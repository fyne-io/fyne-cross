package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	darwinSDKExtractImage  = "fyneio/fyne-cross-images:darwin-sdk-extractor"
	darwinSDKExtractOutDir = "SDKs"
	darwinSDKExtractScript = "darwin-sdk-extractor.sh"
)

// DarwinSDKExtract extracts the macOS SDK from the Command Line Tools for Xcode package
type DarwinSDKExtract struct {
	pull            bool
	sdkPath         string
	containerEngine string
}

// Name returns the one word command name
func (cmd *DarwinSDKExtract) Name() string {
	return "darwin-sdk-extract"
}

// Description returns the command description
func (cmd *DarwinSDKExtract) Description() string {
	return "Extracts the macOS SDK from the Command Line Tools for Xcode package"
}

// Parse parses the arguments and set the usage for the command
func (cmd *DarwinSDKExtract) Parse(args []string) error {
	flagSet.StringVar(&cmd.sdkPath, "xcode-path", "", "Path to the Command Line Tools for Xcode (i.e. /tmp/Command_Line_Tools_for_Xcode_12.5.dmg)")
	// flagSet.StringVar(&cmd.sdkVersion, "sdk-version", "", "SDK version to use. Default to automatic detection")
	flagSet.StringVar(&cmd.containerEngine, "engine", "", "The container engine to use. Supported engines: [docker, podman]. Default to autodetect.")
	flagSet.BoolVar(&cmd.pull, "pull", true, "Attempt to pull a newer version of the docker base image")

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	if cmd.sdkPath == "" {
		return fmt.Errorf("path to the Command Line Tools for Xcode using the 'xcode-path' is required.\nRun 'fyne-cross %s --help' for details", cmd.Name())
	}

	i, err := os.Stat(cmd.sdkPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Command Line Tools for Xcode file %q does not exists", cmd.sdkPath)
	}
	if err != nil {
		return fmt.Errorf("Command Line Tools for Xcode file %q error: %s", cmd.sdkPath, err)
	}
	if i.IsDir() {
		return fmt.Errorf("Command Line Tools for Xcode file %q is a directory", cmd.sdkPath)
	}
	if !strings.HasSuffix(cmd.sdkPath, ".dmg") {
		return fmt.Errorf("Command Line Tools for Xcode file must be in dmg format")
	}

	return nil
}

// Run runs the command
func (cmd *DarwinSDKExtract) Run() error {

	sdkDir := filepath.Dir(cmd.sdkPath)
	dmg := filepath.Base(cmd.sdkPath)
	outDir := filepath.Join(sdkDir, darwinSDKExtractOutDir)

	if _, err := os.Stat(outDir); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("output dir %q already exists. Remove before continue", outDir)
	}

	// mount the fyne-cross volume
	workDir, err := os.MkdirTemp("", cmd.Name())
	if err != nil {
		return err
	}

	vol, err := volume.Mount(workDir, "")
	if err != nil {
		return err
	}

	// attempt to autodetect
	containerEngine, err := MakeEngine(cmd.containerEngine)
	if err != nil {
		return err
	}

	ctx := Context{
		Engine: containerEngine,
		Debug:  true,
		Pull:   cmd.pull,
		Volume: vol,
	}

	engine, err := newLocalContainerEngine(ctx)
	if err != nil {
		return err
	}

	i := engine.createContainerImage("", linuxOS, darwinSDKExtractImage)
	i.SetMount("sdk", sdkDir, "/mnt")
	i.Prepare()

	log.Infof("[i] Extracting SDKs from %q, please wait it could take a while...", dmg)
	err = i.Run(ctx.Volume, options{}, []string{
		darwinSDKExtractScript,
		dmg,
	})
	if err != nil {
		return err
	}
	log.Infof("[✓] SDKs extracted to: %s", outDir)
	return nil
}

// Usage displays the command usage
func (cmd *DarwinSDKExtract) Usage() {
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
