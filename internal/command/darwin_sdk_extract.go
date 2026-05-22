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
				Usage:       "set path to the Command Line Tools for Xcode (i.e. /tmp/Command_Line_Tools_for_Xcode_12.5.dmg)",
				Destination: &cmd.sdkPath,
			},
			&cli.StringFlag{
				Name:        "engine",
				Usage:       "set container engine to use, supported engines: [docker, podman]",
				DefaultText: "autodetect",
				Destination: &cmd.containerEngine,
			},
			&cli.BoolFlag{
				Name:        "pull",
				Usage:       "attempt to pull a newer version of the docker base image",
				Value:       true,
				Destination: &cmd.pull,
			},
		},
		Action: func(ctx *cli.Context) error {
			if cmd.sdkPath == "" {
				return fmt.Errorf("path to the Command Line Tools for Xcode using the 'xcode-path' is required.\nRun 'fyne-cross %s --help' for details", ctx.Command.Name)
			}

			if !strings.HasSuffix(cmd.sdkPath, ".dmg") {
				return fmt.Errorf("Command Line Tools for Xcode file must be in dmg format")
			}

			fi, err := os.Stat(cmd.sdkPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("Command Line Tools for Xcode file %q does not exists", cmd.sdkPath)
			}
			if err != nil {
				return fmt.Errorf("Command Line Tools for Xcode file %q error: %s", cmd.sdkPath, err)
			}
			if fi.IsDir() {
				return fmt.Errorf("Command Line Tools for Xcode file %q is a directory", cmd.sdkPath)
			}

			return cmd.run(ctx)
		},
	}
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
