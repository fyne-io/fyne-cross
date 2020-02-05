package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "develop"
const dockerImage = "lucor/fyne-cross:" + version
const dockerAndroid = "lucor/fyne-cross:android"

// goosWithArch represents the list of supported GOARCH for a GOOS
var goosWithArch = map[string][]string{
	"darwin":  {"amd64", "386"},
	"linux":   {"amd64", "386", "arm", "arm64"},
	"windows": {"amd64", "386"},
	"android": {"amd64", "386", "arm", "arm64"},
}

// targetWithBuildOpts represents the list of supported GOOS/GOARCH with the relative
// options to build
var targetWithBuildOpts = map[string][]string{
	"darwin/amd64":  {"GOOS=darwin", "GOARCH=amd64", "CC=o32-clang"},
	"darwin/386":    {"GOOS=darwin", "GOARCH=386", "CC=o32-clang"},
	"linux/amd64":   {"GOOS=linux", "GOARCH=amd64", "CC=gcc"},
	"linux/386":     {"GOOS=linux", "GOARCH=386", "CC=i686-linux-gnu-gcc"},
	"linux/arm":     {"GOOS=linux", "GOARCH=arm", "CC=arm-linux-gnueabihf-gcc", "GOARM=7"},
	"linux/arm64":   {"GOOS=linux", "GOARCH=arm64", "CC=aarch64-linux-gnu-gcc"},
	"windows/amd64": {"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
	"windows/386":   {"GOOS=windows", "GOARCH=386", "CC=i686-w64-mingw32-gcc"},

	"android/386":   {"GOOS=android", "GOARCH=386", ndk["386"].clangFlag},
	"android/amd64": {"GOOS=android", "GOARCH=amd64", ndk["amd64"].clangFlag},
	"android/arm":   {"GOOS=android", "GOARCH=arm", ndk["arm"].clangFlag, "GOARM=7"},
	"android/arm64": {"GOOS=android", "GOARCH=arm64", ndk["arm64"].clangFlag},
}

// targetLdflags represents the list of default ldflags to pass on build
// for a specified GOOS/GOARCH
var targetLdflags = map[string]string{
	"windows/amd64": "-H windowsgui",
	"windows/386":   "-H windowsgui",
}

// targetTags represents the list of default tags to pass on build
// for a specified GOOS/GOARCH
var targetTags = map[string]string{
	"linux/arm":     "gles",
	"linux/arm64":   "gles",
	"android/386":   "gles",
	"android/amd64": "gles",
	"android/arm":   "gles",
	"android/arm64": "gles",
}

var (
	// targetList represents a list of target to build on separated by comma
	targetList string
	// output represents the named output file
	output string
	// pkg represents the package to build
	pkg string
	// pkgRootDir represents the package root directory
	pkgRootDir string
	// goPath represents the GOPATH to mount into container. It will be used to share/cache sources and dependencies
	goPath string
	// verbosity represents the verbosity setting
	verbose bool
	// ldflags represents the flags to pass to the external linker
	ldflags string
	// printVersion if true will print the fyne-cross version
	printVersion bool
	// noStripDebug if true will not strip debug information from binaries
	noStripDebug bool
	// dist if true will also prepare an application for distribution
	dist bool
	// icon represents the application icon used for distribution. Default to Icon.png
	icon string
)

// builder is the command implementing the fyne app command interface
type builder struct{}

func (b *builder) addFlags() {
	defaultTarget := strings.Join([]string{build.Default.GOOS, build.Default.GOARCH}, "/")
	flag.StringVar(&targetList, "targets", defaultTarget, fmt.Sprintf("The list of targets to build separated by comma. Default to current GOOS/GOARCH %s", defaultTarget))
	flag.StringVar(&output, "output", "", "The named output file. Default to package name")
	flag.StringVar(&pkgRootDir, "dir", "", "The package root directory. Default current dir")
	flag.StringVar(&goPath, "gopath", "", "The local GOPATH to mount into container, used to share/cache sources and dependencies. Default to system cache directory (i.e. $HOME/.cache/fyne-cross)")
	flag.BoolVar(&verbose, "v", false, "Enable verbosity flag for go commands. Default to false")
	flag.StringVar(&ldflags, "ldflags", "", "Flags to pass to the external linker")
	flag.BoolVar(&noStripDebug, "no-strip", false, "If set will not strip debug information from binaries")
	flag.BoolVar(&printVersion, "version", false, "Print fyne-cross version")
	flag.BoolVar(&dist, "dist", false, "If set will also prepare an application for distribution")
	flag.StringVar(&icon, "icon", "Icon.png", "Application icon used for distribution. Default to Icon.png")
}

func (b *builder) printHelp(indent string) {
	fmt.Println("Usage: fyne-cross [parameters] package")
	fmt.Println()
	fmt.Println("Cross compile a Fyne application")
	fmt.Println()

	fmt.Println("Package is the relative path to main.go file or main package. Default to '.'")
	fmt.Println()

	fmt.Println("Optional parameters:")
	flag.PrintDefaults()
	fmt.Println()

	fmt.Println("Supported targets:")
	for target := range targetWithBuildOpts {
		fmt.Println(indent, "- ", target)
	}
	fmt.Println()

	fmt.Println("Default ldflags per target:")
	for target, ldflags := range targetLdflags {
		fmt.Println(indent, "- ", target, ldflags)
	}
	fmt.Println()

	fmt.Println("Example: fyne-cross --targets=linux/amd64,windows/amd64 --output=test ./cmd/test")
}

func (b *builder) run(args []string) {
	var err error

	if printVersion == true {
		fmt.Printf("fyne-cross version %s\n", version)
		os.Exit(2)
	}

	targets, err := parseTargets(targetList)
	if err != nil {
		fmt.Printf("Unable to parse targets option %s", err)
		os.Exit(1)
	}

	if pkgRootDir == "" {
		pkgRootDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Cannot get the path for current directory %s", err)
			os.Exit(1)
		}
	}
	fmt.Println("Project root dir:", pkgRootDir)

	if goPath == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			fmt.Printf("Cannot get the path for the system cache directory %s", err)
			os.Exit(1)
		}
		goPath = filepath.Join(userCacheDir, "fyne-cross")
		err = os.MkdirAll(goPath, 0755)
		if err != nil {
			fmt.Printf("Cannot create the fyne-cross GOPATH under the system cache directory %s", err)
			os.Exit(1)
		}
	}

	pkg := args[0]
	if pkg == "" {
		pkg = "."
	}

	fmt.Println("Package dir:", pkg)

	err = checkRequirements()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// output
	if output == "" {
		// set binary name as package folder dir
		wd := pkg
		if wd == "." {
			wd, err = os.Getwd()
			if err != nil {
				fmt.Printf("Cannot get the path for current directory %s", err)
				os.Exit(1)
			}
		}
		parts := strings.Split(wd, "/")
		output = parts[len(parts)-1]
	}

	fmt.Printf("Build output folder: %s/build\n", pkgRootDir)

	// dist
	if dist {
		fmt.Println("Dist mode enabled", dist)
		fmt.Printf("Icon app: %s\n", icon)
		os.Mkdir("dist", 0755)
		fmt.Printf("Dist output folder: %s/dist\n", pkgRootDir)
	}

	for _, target := range targets {
		fmt.Printf("Target %s\n", target)

		osAndarch := strings.Split(target, "/")
		db := dockerBuilder{
			pkg:        pkg,
			workDir:    pkgRootDir,
			goPath:     goPath,
			output:     output,
			verbose:    verbose,
			ldflags:    ldflags,
			stripDebug: !noStripDebug,
			target:     target,
			os:         osAndarch[0],
			arch:       osAndarch[1],
			dist:       dist,
			icon:       icon,
		}

		t, _ := db.targetOutput()

		// prepare windows resources if dist is specified
		if db.dist && db.os == "windows" {
			err := db.windowsResources()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			exec.Command("cp", fmt.Sprintf("build/%s", t), "dist/").Run()
		}

		// if project does not support modules, download deps with go get
		goModPath := filepath.Join(db.workDir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			fmt.Println("No module found. Creating a temporary one...")
			err = db.goModInit()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Println("Building...")
		err = db.goBuild()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Built as %s\n", t)

		// create dist package, if specified
		if db.dist && (db.os == "linux" || db.os == "darwin") {
			err := db.distPackage(t)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if db.os == "linux" {
				exec.Command("mv", fmt.Sprintf("%s.tar.gz", t), "dist/").Run()
			}
			if db.os == "darwin" {
				exec.Command("mv", fmt.Sprintf("%s.app", t), "dist/").Run()
			}
		}
	}
}

// dockerBuilder represents the docker builder
type dockerBuilder struct {
	output     string
	pkg        string
	workDir    string
	goPath     string
	verbose    bool
	ldflags    string
	stripDebug bool
	target     string
	os         string
	arch       string
	dist       bool
	icon       string
}

// checkRequirements checks if all the build requirements are satisfied
func checkRequirements() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("Missed requirement: docker binary not found in PATH")
	}

	var stderr bytes.Buffer
	cmd := exec.Command("docker", "version")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s", stderr.Bytes())
	}
	return nil
}

// windowsResources creates the windows resources to be linked during build
func (d *dockerBuilder) windowsResources() error {
	args := append(d.defaultArgs(), d.goEnvArgs()...)
	args = append(args, dockerImage)

	cmdArgs := []string{
		"gowindres",
		"-arch", d.arch,
		"-output", d.output,
		"-dir", d.pkg,
	}
	args = append(args, cmdArgs...)

	cmd := exec.Command("docker", args...)
	if d.verbose {
		fmt.Println(cmd)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// distPackage creates the distribution package using fyne package
func (d *dockerBuilder) distPackage(executable string) error {
	args := append(d.defaultArgs(), d.goEnvArgs()...)
	args = append(args, dockerImage)

	cmdArgs := []string{
		"fyne", "package",
		"-os", d.os,
		"-icon", d.icon,
		"-executable", fmt.Sprintf("build/%s", executable),
		"-name", executable,
	}

	args = append(args, cmdArgs...)

	cmd := exec.Command("docker", args...)
	if d.verbose {
		fmt.Println(cmd)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// goModInit downloads the application dependencies via go get
func (d *dockerBuilder) goModInit() error {
	args := append(d.defaultArgs(), d.goEnvArgs()...)
	args = append(args, d.goModInitArgs()...)
	if d.verbose {
		fmt.Printf("docker %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// goBuild runs the go build for target
func (d *dockerBuilder) goBuild() error {
	args := append(d.defaultArgs(), d.goEnvArgs()...)

	buildArgs, err := d.goBuildArgs()
	if err != nil {
		return err
	}

	args = append(args, buildArgs...)
	if d.verbose {
		fmt.Printf("docker %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// targetOutput returns the output file for the specified target.
// Default prefix is the package name. To override use the output option.
// Example: fyne-linux-amd64
func (d *dockerBuilder) targetOutput() (string, error) {
	if d.os == "android" {
		abi := ndk[d.arch].abi
		outputDir := fmt.Sprintf("android/%s", abi)
		return filepath.Join(outputDir, "libfyne.so"), nil
	}

	normalizedTarget := strings.ReplaceAll(d.target, "/", "-")

	ext := ""
	if d.os == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("%s-%s%s", d.output, normalizedTarget, ext), nil
}

// verbosityFlag returns the string used to set verbosity with go commands
// according to current setting
func (d *dockerBuilder) verbosityFlag() string {
	v := ""
	if d.verbose {
		v = "-v"
	}
	return v
}

// defaultArgs returns the default arguments used to run a go command into the
// docker container
func (d *dockerBuilder) defaultArgs() []string {
	args := []string{
		"run",
		"--rm",
		"-t",
	}

	// set workdir
	args = append(args, "-w", fmt.Sprintf("/app"))

	// mount root dir package under /app
	args = append(args, "-v", fmt.Sprintf("%s:/app", d.workDir))

	// mount the cache user dir. Used to cache package dependencies (GOROOT/pkg and GOROOT/src)
	args = append(args, "-v", fmt.Sprintf("%s:/go", d.goPath))

	// attempt to set fyne user id as current user id to handle mount permissions
	// on linux and MacOS
	if runtime.GOOS != "windows" {
		u, err := user.Current()
		if err == nil {
			args = append(args, "-e", fmt.Sprintf("fyne_uid=%s", u.Uid))
		}
	}

	return args
}

// goModInitArgs returns the arguments for the "go get" command
func (d *dockerBuilder) goModInitArgs() []string {
	buildCmd := fmt.Sprintf("go mod init %s", d.output)
	// add docker image
	img := dockerImage
	if d.os == "android" {
		img = dockerAndroid
	}
	return []string{img, buildCmd}
}

func (d *dockerBuilder) goEnvArgs() []string {
	// Start adding env variables
	args := []string{
		// enable CGO
		"-e", "CGO_ENABLED=1",
		// mount GOCACHE to reuse cache between builds
		"-e", "GOCACHE=/go/go-build",
	}

	// add default compile target options env variables
	if buildOpts, ok := targetWithBuildOpts[d.target]; ok {
		for _, o := range buildOpts {
			args = append(args, "-e", o)
		}
	}

	return args
}

// goBuildArgs returns the arguments for the "go build" command for target
func (d *dockerBuilder) goBuildArgs() ([]string, error) {
	args := []string{}

	// add docker image
	img := dockerImage
	if d.os == "android" {
		img = dockerAndroid
	}
	args = append(args, img)

	// add go build command
	args = append(args, "go", "build")

	// Start adding ldflags
	ldflags := []string{}

	// Strip debug information
	if d.stripDebug {
		ldflags = append(ldflags, "-w", "-s")
	}

	// add defaults
	if ldflagsDefault, ok := targetLdflags[d.target]; ok {
		ldflags = append(ldflags, ldflagsDefault)
	}
	// add custom ldflags
	if d.ldflags != "" {
		ldflags = append(ldflags, d.ldflags)
	}

	// add ldflags to command, if any
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", fmt.Sprintf("'%s'", strings.Join(ldflags, " ")))
	}

	// Start adding ldflags
	tags := []string{}

	// add defaults
	if tagsDefault, ok := targetTags[d.target]; ok {
		tags = append(tags, tagsDefault)
	}

	// add tags to command, if any
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")))
	}

	// set c-shared build mode for android
	if d.os == "android" {
		args = append(args, "-buildmode", "c-shared")
	}

	// add target output
	targetOutput, err := d.targetOutput()
	if err != nil {
		return []string{}, err
	}
	args = append(args, "-o", fmt.Sprintf("build/%s", targetOutput))

	// add verbose flag
	if d.verbose {
		args = append(args, "-v")
	}

	// add package
	args = append(args, d.pkg)
	return args, nil
}

// parseTargets parse comma separated target list and validate against the supported targets
func parseTargets(targetList string) ([]string, error) {
	targets := []string{}

	for _, target := range strings.Split(targetList, ",") {
		target = strings.TrimSpace(target)

		osAndArch := strings.Split(target, "/")
		if len(osAndArch) != 2 {
			return targets, fmt.Errorf("Unsupported target %q", target)
		}

		targetOs, targetArch := osAndArch[0], osAndArch[1]
		if targetArch == "*" {
			okArchs, ok := goosWithArch[targetOs]
			if !ok {
				return targets, fmt.Errorf("Unsupported os %q", targetOs)
			}

			for _, arch := range okArchs {
				targets = append(targets, strings.Join([]string{targetOs, arch}, "/"))
			}
			continue
		}

		if _, ok := targetWithBuildOpts[target]; !ok {
			return targets, fmt.Errorf("Unsupported target %q", target)
		}

		targets = append(targets, target)
	}

	return targets, nil
}
