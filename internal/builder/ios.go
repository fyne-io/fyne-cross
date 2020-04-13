package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lucor/fyne-cross/internal/volume"
)

// NewIOS returns a builder for the iOS OS
func NewIOS(arch string, output string) *IOS {
	return &IOS{
		os:     "ios",
		arch:   arch,
		output: output,
	}
}

// IOS is the build for the iOS OS
type IOS struct {
	os     string
	arch   string
	output string
}

// PreBuild performs all tasks needed to perform a build
func (b *IOS) PreBuild(vol *volume.Volume, opts PreBuildOptions) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("iOS compilation is supported only on darwin hosts")
	}
	//ensures go.mod exists, if not try to create a temporary one
	if opts.AppID == "" {
		return fmt.Errorf("appID is required for iOS build")
	}
	return goModInit(vol, opts.Verbose)
}

// Build builds the package
func (b *IOS) Build(vol *volume.Volume, opts BuildOptions) error {
	return nil
}

//BuildEnv returns the env variables required to build the package
func (b *IOS) BuildEnv() []string {
	return []string{}
}

//BuildLdFlags returns the default ldflags used to build the package
func (b *IOS) BuildLdFlags() []string {
	return nil
}

//BuildTags returns the default tags used to build the package
func (b *IOS) BuildTags() []string {
	return nil
}

// TargetID returns the target ID for the builder
func (b *IOS) TargetID() string {
	return fmt.Sprintf("%s", b.os)
}

// Output returns the named output
func (b *IOS) Output() string {
	return b.output
}

// Package generate a package for distribution
func (b *IOS) Package(vol *volume.Volume, opts PackageOptions) error {
	// copy the icon to tmp dir
	err := cp(opts.Icon, volume.JoinPathHost(vol.TmpDirHost(), defaultIcon))
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app due to error copying the icon: %v", err)
	}

	// use the fyne package command to create the dist package
	packageName := b.Output() + ".app"
	command := []string{
		fyneCmd, "package",
		"-os", b.os,
		"-name", b.Output(),
		"-icon", volume.JoinPathContainer(vol.TmpDirContainer(), defaultIcon),
		"-appID", opts.AppID, // opts.AppID is mandatory for iOS app
	}

	err = dockerCmd(iosDockerImage, vol, []string{}, vol.WorkDirContainer(), command, opts.Verbose).Run()
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := volume.JoinPathHost(vol.WorkDirHost(), packageName)
	distFile := volume.JoinPathHost(vol.DistDirHost(), b.TargetID(), packageName)
	err = os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("Could not create the dist package dir: %v", err)
	}
	return os.Rename(srcFile, distFile)
}
