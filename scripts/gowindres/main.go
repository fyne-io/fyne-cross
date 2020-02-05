package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"path"
	"text/template"

	ico "github.com/Kodeworks/golang-image-ico"
)

var (
	// output represents the named output file
	output string
	// pkgRootDir represents the package root directory
	pkgRootDir string
	// icon represents the application icon used for distribution. Default to Icon.png
	icon string
	// the architecture: 386 or amd64
	arch string
)

const (
	rc          = "main.rc"
	target386   = "pe-i386"
	targetAmd64 = "pe-x86-64"
	windresBin  = "x86_64-w64-mingw32-windres"
)

type tplData struct {
	Name string
}

func main() {

	flag.StringVar(&icon, "icon", "Icon.png", "Application icon used for distribution. Default to Icon.png")
	flag.StringVar(&arch, "arch", "", "Architecture: 368 or amd64")
	flag.StringVar(&output, "output", "", "The named output file")
	flag.StringVar(&pkgRootDir, "dir", "", "The package root directory")
	flag.Parse()

	manifestPath := path.Join(pkgRootDir, output+".exe.manifest")
	rcPath := path.Join(pkgRootDir, rc)
	resource := output + ".syso"

	// convert png to ico
	pngPath := icon
	if icon == "Icon.png" {
		pngPath = path.Join(pkgRootDir, "Icon.png")
	}
	icoPath := path.Join(pkgRootDir, output+".ico")
	err := convertPngToIco(pngPath, icoPath)
	if err != nil {
		log.Fatal(err)
	}

	data := tplData{
		Name: output,
	}

	// write the main.rc
	err = writeRc(data, rcPath)
	if err != nil {
		log.Fatal(err)
	}

	// write the manifest
	err = writeManifest(data, manifestPath)
	if err != nil {
		log.Fatal(err)
	}

	target := targetAmd64
	if arch == "386" {
		target = target386
	}

	cmd := exec.Command(windresBin, "-F", target, "-o", resource, rc)
	cmd.Dir = pkgRootDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print("Could not create windows resource", err)
		fmt.Printf("Debug output: %s\n", out)
		os.Exit(1)
	}
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

// writeManifest writes the manifest
func writeManifest(data tplData, outFile string) error {

	manifest := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
  <assemblyIdentity version="1.0.0.0" processorArchitecture="*" name="{{.Name}}" type="win32"/>
</assembly>"`

	mFile, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("Could not create the manifest file: %s", err)
	}
	defer mFile.Close()
	tpl, err := template.New("tpl").Parse(manifest)
	if err != nil {
		return fmt.Errorf("Could not parse the manifest template: %s", err)
	}
	err = tpl.Execute(mFile, data)
	if err != nil {
		return fmt.Errorf("Could not execute the manifest template: %s", err)
	}
	return nil
}

// writeRs writes the rc file
func writeRc(data tplData, outFile string) error {

	rc := `100 ICON    "{{.Name}}.ico"
100 24      "{{.Name}}.exe.manifest"
`
	mFile, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("Could not create the rc file: %s", err)
	}
	defer mFile.Close()
	tpl, err := template.New("tpl").Parse(rc)
	if err != nil {
		return fmt.Errorf("Could not parse the rc template: %s", err)
	}
	err = tpl.Execute(mFile, data)
	if err != nil {
		return fmt.Errorf("Could not execute the rc template: %s", err)
	}
	return nil
}
