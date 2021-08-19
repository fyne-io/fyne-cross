/*
gowindres is an utility used internally by fyne-cross to generate windows resource
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"text/template"

	"golang.org/x/sys/execabs"
)

var (
	// output represents the named output file
	output string
	// workDir represents the package root directory
	workDir string
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

	flag.StringVar(&arch, "arch", "", "Architecture: 368 or amd64")
	flag.StringVar(&output, "output", "", "The named output file")
	flag.StringVar(&workDir, "workdir", "", "The working directory")
	flag.Parse()

	manifestPath := path.Join(workDir, output+".manifest")
	rcPath := path.Join(workDir, rc)
	resource := output + ".syso"

	data := tplData{
		Name: output,
	}

	// write the main.rc
	err := writeRc(data, rcPath)
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

	cmd := execabs.Command(windresBin, "-F", target, "-o", resource, rc)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Could not create windows resource", err)
		fmt.Printf("Debug output: %s\n", out)
		os.Exit(1)
	}
}

// writeManifest writes the manifest
func writeManifest(data tplData, outFile string) error {

	manifest := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
  <assemblyIdentity version="1.0.0.0" processorArchitecture="*" name="{{.Name}}" type="win32"/>
</assembly>"`

	mFile, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("could not create the manifest file: %s", err)
	}
	defer mFile.Close()
	tpl, err := template.New("tpl").Parse(manifest)
	if err != nil {
		return fmt.Errorf("could not parse the manifest template: %s", err)
	}
	err = tpl.Execute(mFile, data)
	if err != nil {
		return fmt.Errorf("could not execute the manifest template: %s", err)
	}
	return nil
}

// writeRs writes the rc file
func writeRc(data tplData, outFile string) error {

	rc := `100 ICON    "{{.Name}}.ico"
100 24      "{{.Name}}.manifest"
`
	mFile, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("could not create the rc file: %s", err)
	}
	defer mFile.Close()
	tpl, err := template.New("tpl").Parse(rc)
	if err != nil {
		return fmt.Errorf("could not parse the rc template: %s", err)
	}
	err = tpl.Execute(mFile, data)
	if err != nil {
		return fmt.Errorf("could not execute the rc template: %s", err)
	}
	return nil
}
