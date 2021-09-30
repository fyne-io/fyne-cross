package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/resource"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/sys/execabs"
)

// DarwinImage builds the darwin docker image
type DarwinImage struct {
	sdkPath    string
	sdkVersion string
}

// Name returns the one word command name
func (cmd *DarwinImage) Name() string {
	return "darwin-image"
}

// Description returns the command description
func (cmd *DarwinImage) Description() string {
	return "Build the docker image for darwin"
}

// Parse parses the arguments and set the usage for the command
func (cmd *DarwinImage) Parse(args []string) error {
	flagSet.StringVar(&cmd.sdkPath, "xcode-path", "", "Path to the Command Line Tools for Xcode (i.e. /tmp/Command_Line_Tools_for_Xcode_12.5.dmg)")
	flagSet.StringVar(&cmd.sdkVersion, "sdk-version", "", "SDK version to use. Default to automatic detection")

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	if cmd.sdkPath == "" {
		return errors.New("path to the Command Line Tools for Xcode using the 'xcode-path' is required.\nRun 'fyne-cross darwin-image --help' for details")
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
func (cmd *DarwinImage) Run() error {

	workDir, err := ioutil.TempDir(os.TempDir(), "fyne-cross-darwin-build")
	if err != nil {
		return fmt.Errorf("could not create temporary dir: %s", err)
	}
	defer os.RemoveAll(workDir)

	log.Info("[i] Building docker darwin image...")
	log.Infof("[i] Work dir: %s", workDir)

	xcodeFile := volume.JoinPathHost(workDir, "command_line_tools_for_xcode.dmg")
	log.Infof("[i] Copying the Command Line Tools for Xcode from %s to %s...", cmd.sdkPath, xcodeFile)
	err = volume.Copy(cmd.sdkPath, xcodeFile)
	if err != nil {
		return fmt.Errorf("could not copy the Command Line Tools for Xcode file into the work dir: %s", err)
	}
	log.Infof("[✓] Command Line Tools for Xcode copied")

	err = ioutil.WriteFile(volume.JoinPathHost(workDir, "Dockerfile"), []byte(resource.DockerfileDarwin), 0644)
	if err != nil {
		return fmt.Errorf("could not create the Dockerfile into the work dir: %s", err)
	}
	log.Infof("[✓] Dockerfile created")

	log.Info("[i] Building docker image...")
	ver := "auto"
	if cmd.sdkVersion != "" {
		ver = cmd.sdkVersion
	}
	log.Info("[i] macOS SDK: ", ver)

	// run the command from the host
	dockerCmd := execabs.Command("docker", "build", "--pull", "--build-arg", fmt.Sprintf("SDK_VERSION=%s", cmd.sdkVersion), "-t", darwinImage, ".")
	dockerCmd.Dir = workDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	err = dockerCmd.Run()
	if err != nil {
		return fmt.Errorf("could not create the docker darwin image: %v", err)
	}
	log.Infof("[✓] Docker image created: %s", darwinImage)
	return nil
}

// Usage displays the command usage
func (cmd *DarwinImage) Usage() {
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
