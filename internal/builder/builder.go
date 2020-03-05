/*
Package builder implements the build actions for the supperted OS and arch
*/
package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lucor/fyne-cross/internal/volume"
)

const (
	baseDockerImage    = "lucor/fyne-cross:develop"
	linuxDockerImage   = baseDockerImage
	windowsDockerImage = baseDockerImage
	darwinDockerImage  = baseDockerImage

	defaultIcon = "Icon.png"
)

// Builder represents a builder
type Builder interface {
	PreBuild(vol *volume.Volume, opts PreBuildOptions) error
	Build(vol *volume.Volume, opts BuildOptions) error
	BuildEnv() []string
	BuildLdFlags() []string
	BuildTags() []string
	Package(vol *volume.Volume, opts PackageOptions) error
	Output() string
	TargetID() string
}

// PreBuildOptions holds the options for the pre build step
type PreBuildOptions struct {
	Verbose bool   // Verbose if true, enable verbose mode
	Icon    string // Icon is the optional icon in png format to use for distribution
}

// BuildOptions holds the options to build the package
type BuildOptions struct {
	Package    string   // Package is the package to build named by the import path as per 'go build'
	LdFlags    []string // LdFlags are the ldflags to pass to the compiler
	Tags       []string // Tags are the tags to pass to the compiler
	StripDebug bool     // StripDebug if true, strips binary output
	Verbose    bool     // Verbose if true, enable verbose mode
}

// PackageOptions holds the options to generate a package for distribution
type PackageOptions struct {
	Icon    string // Icon is the optional icon in png format to use for distribution
	AppID   string // Icon is the appID to use for distribution
	Verbose bool   // Verbose if true, enable verbose mode
}

// dockerCmd exec a command inside the container for the specified image
func dockerCmd(image string, vol *volume.Volume, env []string, workDir string, command []string, verbose bool) *exec.Cmd {
	// define workdir
	w := vol.WorkDirContainer()
	if workDir != "" {
		w = workDir
	}

	args := []string{
		"run", "--rm", "-t",
		"-w", w, // set workdir
		"-v", fmt.Sprintf("%s:%s", vol.WorkDirHost(), vol.WorkDirContainer()), // mount the working dir
		"-v", fmt.Sprintf("%s:%s", vol.CacheDirHost(), vol.CacheDirContainer()), // mount the cache dir
		"-e", "CGO_ENABLED=1", // enable CGO
		"-e", fmt.Sprintf("GOCACHE=%s", vol.GoCacheDirContainer()), // mount GOCACHE to reuse cache between builds
	}

	// add custom env variables
	for _, e := range env {
		args = append(args, "-e", e)
	}

	// attempt to set fyne user id as current user id to handle mount permissions
	// on linux and MacOS
	if runtime.GOOS != "windows" {
		u, err := user.Current()
		if err == nil {
			args = append(args, "-e", fmt.Sprintf("fyne_uid=%s", u.Uid))
		}
	}

	// specify the image to use
	args = append(args, image)

	// add the command to execute
	args = append(args, command...)

	// run the command inside the container
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if verbose {
		fmt.Println(cmd.String())
	}

	return cmd
}

// goModInit ensure a go.mod exists. If not try to generates a temporary one
func goModInit(vol *volume.Volume, verbose bool) error {
	// check if the go.mod exists
	goModPath := filepath.Join(vol.WorkDirHost(), "go.mod")
	_, err := os.Stat(goModPath)
	if err == nil {
		if verbose {
			fmt.Println("go.mod found")
		}
		return nil
	}

	if verbose {
		fmt.Println("go.mod not found, creating a temporary one...")
	}

	// Module does not exists, generate a temporary one
	command := "go mod init fyne-cross-temp-module"
	err = dockerCmd(baseDockerImage, vol, []string{}, vol.WorkDirContainer(), []string{command}, verbose).Run()

	if err != nil {
		return fmt.Errorf("Could not generate the temporary go module: %v", err)
	}
	return nil
}

// goBuildCmd returns the go build command
func goBuildCmd(output string, opts BuildOptions) []string {
	// add go build command
	args := []string{"go", "build"}

	ldflags := opts.LdFlags
	// Strip debug information
	if opts.StripDebug {
		ldflags = append(ldflags, "-w", "-s")
	}

	// add ldflags to command, if any
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", fmt.Sprintf("'%s'", strings.Join(ldflags, " ")))
	}

	// add tags to command, if any
	tags := opts.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")))
	}

	args = append(args, "-o", output)

	// add verbose flag
	if opts.Verbose {
		args = append(args, "-v")
	}

	//add package
	args = append(args, opts.Package)
	return args
}

// cp is copies a resource from src to dest
func cp(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
