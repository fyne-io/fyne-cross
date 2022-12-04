package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const (
	// fyneBin is the path of the fyne binary into the docker image
	fyneBin = "/usr/local/bin/fyne"
	// gowindresBin is the path of the gowindres binary into the docker image
	gowindresBin = "/usr/local/bin/gowindres"
)

type containerEngine interface {
	createContainerImage(arch Architecture, OS string, image string) containerImage
}

type baseEngine struct {
	containerEngine

	env  map[string]string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	tags []string          // Tags defines the tags to use

	vol volume.Volume
}

type containerImage interface {
	Run(vol volume.Volume, opts options, cmdArgs []string) error
	Prepare() error
	Finalize(packageName string) error

	Architecture() Architecture
	OS() string
	ID() string
	Target() string
	Env(string) (string, bool)
	SetEnv(string, string)
	UnsetEnv(string)
	AllEnv() []string
	SetMount(string, string, string)
	AppendTag(string)
	Tags() []string

	Engine() containerEngine
}

type containerMountPoint struct {
	name        string
	localHost   string
	inContainer string
}

type baseContainerImage struct {
	arch Architecture // Arch defines the target architecture
	os   string       // OS defines the target OS
	id   string       // ID is the context ID

	env   map[string]string     // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	tags  []string              // Tags defines the tags to use
	mount []containerMountPoint // Mount point between local host [key] and in container point [target]

	DockerImage string // DockerImage defines the docker image used to build
}

func newContainerEngine(context Context) (containerEngine, error) {
	if context.Engine.IsDocker() || context.Engine.IsPodman() {
		return newLocalContainerEngine(context)
	}
	if context.Engine.IsKubernetes() {
		return newKubernetesContainerRunner(context)
	}
	return nil, fmt.Errorf("unknown engine: '%s'", context.Engine)
}

var debugEnable bool

func debugging() bool {
	return debugEnable
}

func (a *baseEngine) createContainerImageInternal(arch Architecture, OS string, image string, fn func(base baseContainerImage) containerImage) containerImage {
	var ID string

	if arch == "" || arch == ArchMultiple {
		ID = OS
	} else {
		ID = fmt.Sprintf("%s-%s", OS, arch)
	}

	ret := fn(baseContainerImage{arch: arch, os: OS, id: ID, DockerImage: image, env: make(map[string]string), tags: a.tags})

	// mount the working dir
	ret.SetMount("project", a.vol.WorkDirHost(), a.vol.WorkDirContainer())

	return ret
}

func (a *baseContainerImage) Architecture() Architecture {
	return a.arch
}

func (a *baseContainerImage) OS() string {
	return a.os
}

func (a *baseContainerImage) ID() string {
	return a.id
}

func (a *baseContainerImage) Target() string {
	target := a.OS()
	if target == androidOS && a.Architecture() != ArchMultiple {
		target += "/" + a.Architecture().String()
	}

	return target
}

func (a *baseContainerImage) Env(key string) (v string, ok bool) {
	v, ok = a.env[key]
	return
}

func (a *baseContainerImage) SetEnv(key string, value string) {
	a.env[key] = value
}

func (a *baseContainerImage) UnsetEnv(key string) {
	delete(a.env, key)
}

func (a *baseContainerImage) AllEnv() []string {
	r := []string{}

	for key, value := range a.env {
		r = append(r, key+"="+value)
	}
	return r
}

func (a *baseContainerImage) SetMount(name string, local string, inContainer string) {
	a.mount = append(a.mount, containerMountPoint{name: name, localHost: local, inContainer: inContainer})
}

func (a *baseContainerImage) AppendTag(tag string) {
	a.tags = append(a.tags, tag)
}

func (a *baseContainerImage) Tags() []string {
	return a.tags
}

// goModInit ensure a go.mod exists. If not try to generates a temporary one
func goModInit(ctx Context, image containerImage) error {
	if ctx.NoProjectUpload {
		return nil
	}

	goModPath := volume.JoinPathHost(ctx.WorkDirHost(), "go.mod")
	log.Infof("[i] Checking for go.mod: %s", goModPath)
	_, err := os.Stat(goModPath)
	if err == nil {
		log.Info("[✓] go.mod found")
		return nil
	}

	log.Info("[i] go.mod not found, creating a temporary one...")
	err = image.Run(ctx.Volume, options{}, []string{"go", "mod", "init", ctx.Name})
	if err != nil {
		return fmt.Errorf("could not generate the temporary go module: %v", err)
	}

	log.Info("[✓] go.mod created")
	return nil
}

// goBuild run the go build command in the container
func goBuild(ctx Context, image containerImage) error {
	log.Infof("[i] Building binary...")
	// add go build command
	args := []string{"go", "build", "-trimpath"}

	// Strip debug information
	if ctx.StripDebug {
		// ensure that CGO_LDFLAGS is not overwritten as they can be passed
		// by the --env argument
		if v, ok := image.Env("CGO_LDFLAGS"); ok {
			// append the ldflags after the existing CGO_LDFLAGS
			image.SetEnv("CGO_LDFLAGS", strings.Trim(v, "\"")+" -w -s")
		} else {
			// create the CGO_LDFLAGS environment variable and add the strip debug flags
			image.SetEnv("CGO_LDFLAGS", "-w -s")
		}
	}

	ldflags := ctx.LdFlags
	// honour the GOFLAGS env variable adding to existing ones
	if v, ok := image.Env("GOFLAGS"); ok {
		ldflags = append(ldflags, v)
	}

	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}

	// add tags to command, if any
	tags := image.Tags()
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}

	// set output folder to fyne-cross/bin/<target>
	binaryName := ctx.Name
	if image.OS() == darwinOS {
		// replicate how fyne package names the binary
		binaryName = calculateExeName(volume.JoinPathHost(ctx.WorkDirHost()), image.OS())
	}
	output := volume.JoinPathContainer(ctx.Volume.BinDirContainer(), image.ID(), binaryName)

	args = append(args, "-o", output)

	// enable debug mode
	if debugging() {
		args = append(args, "-v")
	}

	//add package
	args = append(args, ctx.Package)

	err := image.Run(ctx.Volume, options{}, args)

	if err != nil {
		return err
	}

	log.Infof("[✓] Binary: %s", volume.JoinPathHost(ctx.BinDirHost(), image.ID(), ctx.Name))
	return nil
}

// fynePackage package the application using the fyne cli tool
func fynePackage(ctx Context, image containerImage) error {
	if debugging() {
		err := image.Run(ctx.Volume, options{}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
	}

	target := image.Target()

	args := []string{
		fyneBin, "package",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := image.Tags()
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	// Enable release mode, if specified
	if ctx.Release {
		args = append(args, "-release")
	}

	// workDir default value
	workDir := ctx.WorkDirContainer()

	if image.OS() == androidOS || image.OS() == webOS {
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
	}

	// linux, darwin and freebsd targets are built by fyne-cross
	// in these cases fyne tool is used only to package the app specifying the executable flag
	if image.OS() == linuxOS || image.OS() == darwinOS || image.OS() == freebsdOS {
		binaryName := ctx.Name
		if image.OS() == darwinOS {
			// replicate how fyne package names the binary
			binaryName = calculateExeName(volume.JoinPathHost(ctx.WorkDirHost()), image.OS())
		}

		args = append(args, "-executable", volume.JoinPathContainer(ctx.BinDirContainer(), image.ID(), binaryName))
		workDir = volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID())
	}

	runOpts := options{
		WorkDir: workDir,
	}

	err := image.Run(ctx.Volume, runOpts, args)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// calculateExeName is ported from the fyne base code to ensure darwin binary naming is consistent between fyne-cross and fyne package
func calculateExeName(sourceDir, os string) string {
	exeName := filepath.Base(sourceDir)
	/* #nosec */
	if data, err := ioutil.ReadFile(filepath.Join(sourceDir, "go.mod")); err == nil {
		modulePath := modfile.ModulePath(data)
		moduleName, _, ok := module.SplitPathVersion(modulePath)
		if ok {
			paths := strings.Split(moduleName, "/")
			name := paths[len(paths)-1]
			if name != "" {
				exeName = name
			}
		}
	}

	if os == windowsOS {
		exeName = exeName + ".exe"
	}

	return exeName
}

// fyneRelease package and release the application using the fyne cli tool
// Note: at the moment this is used only for the android builds
func fyneRelease(ctx Context, image containerImage) error {
	if debugging() {
		err := image.Run(ctx.Volume, options{}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
		return nil
	}

	target := image.Target()

	args := []string{
		fyneBin, "release",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := image.Tags()
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	// workDir default value
	workDir := ctx.WorkDirContainer()

	switch image.OS() {
	case androidOS:
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
		if ctx.Keystore != "" {
			args = append(args, "-keyStore", ctx.Keystore)
		}
		if ctx.KeystorePass != "" {
			args = append(args, "-keyStorePass", ctx.KeystorePass)
		}
		if ctx.KeyPass != "" {
			args = append(args, "-keyPass", ctx.KeyPass)
		}
	case iosOS:
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Profile != "" {
			args = append(args, "-profile", ctx.Profile)
		}
	case webOS:
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
	case windowsOS:
		if ctx.Certificate != "" {
			args = append(args, "-certificate", ctx.Certificate)
		}
		if ctx.Developer != "" {
			args = append(args, "-developer", ctx.Developer)
		}
		if ctx.Password != "" {
			args = append(args, "-password", ctx.Password)
		}
	}

	runOpts := options{
		WorkDir: workDir,
	}

	err := image.Run(ctx.Volume, runOpts, args)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// WindowsResource create a windows resource under the project root
// that will be automatically linked by compliler during the build
func WindowsResource(ctx Context, image containerImage) (string, error) {
	args := []string{
		gowindresBin,
		"-arch", image.Architecture().String(),
		"-output", ctx.Name,
		"-workdir", volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID()),
	}

	runOpts := options{
		WorkDir: volume.JoinPathContainer(ctx.TmpDirContainer(), image.ID()),
	}

	err := image.Run(ctx.Volume, runOpts, args)
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// copy the windows resource under the package root
	// it will be automatically linked by compiler during build
	windres := ctx.Name + ".syso"
	out := filepath.Join(ctx.Package, windres)
	err = volume.Copy(volume.JoinPathHost(ctx.TmpDirHost(), image.ID(), windres), volume.JoinPathHost(ctx.WorkDirHost(), out))
	if err != nil {
		return "", fmt.Errorf("could not copy windows resource under the package root: %v", err)
	}

	return out, nil
}
