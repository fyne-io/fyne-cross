package builder

import (
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/lucor/fyne-cross/internal/volume"
)

// NewWindows returns a builder for the Windows OS
func NewWindows(arch string, output string) *Windows {
	return &Windows{
		os:     "windows",
		arch:   arch,
		output: output,
	}
}

// Windows is the build for the Windows OS
type Windows struct {
	os     string
	arch   string
	output string
}

// PreBuild performs all tasks needed to perform a build
func (b *Windows) PreBuild(vol *volume.Volume, opts PreBuildOptions) error {
	// ensures go.mod exists, if not try to create a temporary one
	err := goModInit(vol, opts.Verbose)
	if err != nil {
		return err
	}

	// Convert the png icon to ico format and store under the temp directory
	convertPngToIco(opts.Icon, path.Join(vol.TmpDirHost(), b.output+".ico"))

	// use the gowindres command to create the windows resource
	command := []string{
		"gowindres",
		"-arch", b.arch,
		"-output", b.output,
		"-workdir", vol.TmpDirContainer(),
	}

	windres := b.output + ".syso"
	err = dockerCmd(windowsDockerImage, vol, []string{}, vol.TmpDirContainer(), command, opts.Verbose).Run()
	if err != nil {
		return fmt.Errorf("Could not create the windows resource %q: %v", windres, err)
	}

	// copy the windows resource under the project root
	// it will be automatically linked by compliler during build
	err = cp(filepath.Join(vol.TmpDirHost(), windres), filepath.Join(vol.WorkDirHost(), windres))
	if err != nil {
		return fmt.Errorf("Could not copy windows resource under the project root: %v", err)
	}
	return nil
}

// Build builds the package
func (b *Windows) Build(vol *volume.Volume, opts BuildOptions) error {

	output := filepath.Join(vol.BinDirContainer(), b.TargetID(), b.Output())

	// add default ldflags, if any
	if ldflags := b.BuildLdFlags(); ldflags != nil {
		opts.LdFlags = append(opts.LdFlags, ldflags...)
	}

	// add default tags, if any
	if tags := b.BuildTags(); tags != nil {
		opts.Tags = append(opts.Tags, tags...)
	}

	command := goBuildCmd(output, opts)
	err := dockerCmd(windowsDockerImage, vol, b.BuildEnv(), vol.WorkDirContainer(), command, opts.Verbose).Run()
	if err != nil {
		return fmt.Errorf("Could not build for %s/%s: %v", b.os, b.arch, err)
	}

	return nil
}

//BuildEnv returns the env variables required to build the package
func (b *Windows) BuildEnv() []string {
	switch b.arch {
	case "amd64":
		return []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"}
	case "386":
		return []string{"GOOS=windows", "GOARCH=386", "CC=i686-w64-mingw32-gcc"}
	}
	return []string{}
}

//BuildLdFlags returns the default ldflags used to build the package
func (b *Windows) BuildLdFlags() []string {
	return []string{"-H windowsgui"}
}

//BuildTags returns the default tags used to build the package
func (b *Windows) BuildTags() []string {
	return nil
}

// TargetID returns the target ID for the builder
func (b *Windows) TargetID() string {
	return fmt.Sprintf("%s-%s", b.os, b.arch)
}

// Output returns the named output
func (b *Windows) Output() string {
	return b.output + ".exe"
}

// WindresOutput returns the named output for the windows resource
func (b *Windows) windresOutput() string {
	return fmt.Sprintf("%s.syso", b.output)
}

// Package generate a package for distribution
func (b *Windows) Package(vol *volume.Volume, opts PackageOptions) error {
	os.Remove(filepath.Join(vol.WorkDirHost(), b.windresOutput()))
	// move the dist package into the "dist" folder
	srcFile := filepath.Join(vol.BinDirHost(), b.TargetID(), b.Output())
	distFile := filepath.Join(vol.DistDirHost(), b.TargetID(), b.Output())
	err := os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("Could not create the dist package dir: %v", err)
	}
	return cp(srcFile, distFile)
}

func convertPngToIco(pngPath string, icoPath string) error {
	// convert icon
	img, err := os.Open(pngPath)
	if err != nil {
		return fmt.Errorf("Failed to open source image: %s", err)
	}
	defer img.Close()
	srcImg, _, err := image.Decode(img)
	if err != nil {
		return fmt.Errorf("Failed to decode source image: %s", err)
	}

	file, err := os.Create(icoPath)
	if err != nil {
		return fmt.Errorf("Failed to open image file: %s", err)
	}
	defer file.Close()
	err = ico.Encode(file, srcImg)
	if err != nil {
		return fmt.Errorf("Failed to write image file: %s", err)
	}
	return nil
}
