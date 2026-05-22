package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

const (
	darwinSDKExtractImage  = "fyneio/fyne-cross-images:darwin-sdk-extractor"
	darwinSDKExtractOutDir = "SDKs"
	darwinSDKExtractScript = "darwin-sdk-extractor.sh"
)

// DarwinSDKExtract extracts the macOS SDK from the Command Line Tools for Xcode package
type darwinSDKExtract struct {
	pull            bool
	sdkPath         string
	containerEngine string
}

// Name returns the one word command name
func DarwinSDKExtract() *cli.Command {
	cmd := &darwinSDKExtract{}
	return &cli.Command{
		Name:  "darwin-sdk-extract",
		Usage: "Extracts the macOS SDK from the Command Line Tools for Xcode package",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "xcode-path",
				Usage:       "Path to the Command Line Tools for Xcode (i.e. /tmp/Command_Line_Tools_for_Xcode_12.5.dmg)",
				Destination: &cmd.sdkPath,
			},
			&cli.StringFlag{
				Name:        "engine",
				Usage:       "The container engine to use. Supported engines: [docker, podman]. Default to autodetect.",
				Destination: &cmd.containerEngine,
			},
			&cli.BoolFlag{
				Name:        "pull",
				Usage:       "Attempt to pull a newer version of the docker base image",
				Value:       true,
				Destination: &cmd.pull,
			},
		},
		Action: func(ctx *cli.Context) error {
			if err := cmd.parse(ctx); err != nil {
				return err
			}
			return cmd.run(ctx)
		},
	}
}

// Parse parses the arguments and set the usage for the command
func (cmd *darwinSDKExtract) parse(ctx *cli.Context) error {
	cmd.sdkPath = ctx.String("xcode-path")
	cmd.containerEngine = ctx.String("engine")
	cmd.pull = ctx.Bool("pull")

	if cmd.sdkPath == "" {
		return fmt.Errorf("path to the Command Line Tools for Xcode using the 'xcode-path' is required.\nRun 'fyne-cross %s --help' for details", ctx.Command.Name)
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
func (cmd *darwinSDKExtract) run(cCtx *cli.Context) error {
	sdkDir := filepath.Dir(cmd.sdkPath)
	dmg := filepath.Base(cmd.sdkPath)
	outDir := filepath.Join(sdkDir, darwinSDKExtractOutDir)

	if _, err := os.Stat(outDir); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("output dir %q already exists. Remove before continue", outDir)
	}

	// mount the fyne-cross volume
	workDir, err := os.MkdirTemp("", cCtx.Command.Name)
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
