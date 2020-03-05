package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucor/fyne-cross/internal/volume"
)

// NewAndroid returns a builder for the linux OS
func NewAndroid(arch string, output string) *Android {
	return &Android{
		os:     "android",
		arch:   arch,
		output: output,
	}
}

// Android is the build for the linux OS
type Android struct {
	os     string
	arch   string
	output string
}

// PreBuild performs all tasks needed to perform a build
func (b *Android) PreBuild(vol *volume.Volume, opts PreBuildOptions) error {
	//ensures go.mod exists, if not try to create a temporary one
	if opts.AppID == "" {
		return fmt.Errorf("appID is required for android build")
	}
	return goModInit(vol, opts.Verbose)
}

// Build builds the package
func (b *Android) Build(vol *volume.Volume, opts BuildOptions) error {
	return nil
}

//BuildEnv returns the env variables required to build the package
func (b *Android) BuildEnv() []string {
	return []string{}
}

//BuildLdFlags returns the default ldflags used to build the package
func (b *Android) BuildLdFlags() []string {
	return nil
}

//BuildTags returns the default tags used to build the package
func (b *Android) BuildTags() []string {
	return nil
}

// TargetID returns the target ID for the builder
func (b *Android) TargetID() string {
	return fmt.Sprintf("%s", b.os)
}

// Output returns the named output
func (b *Android) Output() string {
	return b.output
}

// Package generate a package for distribution
func (b *Android) Package(vol *volume.Volume, opts PackageOptions) error {
	// copy the icon to tmp dir
	err := cp(opts.Icon, filepath.Join(vol.TmpDirHost(), defaultIcon))
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app due to error copying the icon: %v", err)
	}

	// use the fyne package command to create the dist package
	packageName := b.Output() + ".apk"
	command := []string{
		fyneCmd, "package",
		"-os", b.os,
		"-name", b.Output(),
		"-icon", filepath.Join(vol.TmpDirContainer(), defaultIcon),
		"-appID", opts.AppID, // opts.AppID is mandatory for android app
	}

	err = dockerCmd(androidDockerImage, vol, []string{}, vol.WorkDirContainer(), command, opts.Verbose).Run()
	if err != nil {
		return fmt.Errorf("Could not package the Fyne app: %v", err)
	}

	// move the dist package into the "dist" folder
	srcFile := filepath.Join(vol.WorkDirHost(), packageName)
	distFile := filepath.Join(vol.DistDirHost(), b.TargetID(), packageName)
	err = os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("Could not create the dist package dir: %v", err)
	}
	return os.Rename(srcFile, distFile)
}
