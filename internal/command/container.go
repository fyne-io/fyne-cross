package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/icon"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/sys/execabs"
)

const (
	// fyneBin is the path of the fyne binary into the docker image
	fyneBin = "/usr/local/bin/fyne"
	// gowindresBin is the path of the gowindres binary into the docker image
	gowindresBin = "/usr/local/bin/gowindres"
)

type ContainerRunner interface {
	NewImageContainer(arch Architecture, OS string, image string) ContainerImage

	Debug(v ...interface{})
	GetDebug() bool
}

type AllContainerRunner struct {
	ContainerRunner

	Env  map[string]string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	Tags []string          // Tags defines the tags to use

	vol volume.Volume

	debug bool
}

type ContainerImage interface {
	Cmd(vol volume.Volume, opts Options, cmdArgs []string) *execabs.Cmd
	Run(vol volume.Volume, opts Options, cmdArgs []string) error
	Prepare() error
	Finalize(srcFile string, packageName string) error

	GetArchitecture() Architecture
	GetOS() string
	GetID() string
	GetTarget() string
	GetEnv(string) (string, bool)
	SetEnv(string, string)
	SetMount(string, string)
	AppendTag(string)

	GetRunner() ContainerRunner
}

type AllContainerImage struct {
	Architecture        // Arch defines the target architecture
	OS           string // OS defines the target OS
	ID           string // ID is the context ID

	Env   map[string]string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	Tags  []string          // Tags defines the tags to use
	Mount map[string]string // Mount point between local host [key] and in container point [target]

	DockerImage string // DockerImage defines the docker image used to build
}

func NewContainerRunner(context Context) ContainerRunner {
	if context.Engine.IsDocker() || context.Engine.IsPodman() {
		return NewLocalContainerRunner(context)
	}
	return nil
}

func (a *AllContainerRunner) Debug(v ...interface{}) {
	if a.debug {
		log.Debug(v...)
	}
}

func (a *AllContainerRunner) GetDebug() bool {
	return a.debug
}

func (a *AllContainerRunner) newImageContainerInternal(arch Architecture, OS string, image string, fn func(arch Architecture, OS string, ID string, image string) ContainerImage) ContainerImage {
	var ID string

	if arch == "" || arch == ArchMultiple {
		ID = OS
	} else {
		ID = fmt.Sprintf("%s-%s", OS, arch)
	}

	ret := fn(arch, OS, ID, image)

	// mount the working dir
	ret.SetMount(a.vol.WorkDirHost(), a.vol.WorkDirContainer())

	return ret
}

func (a *AllContainerImage) GetArchitecture() Architecture {
	return a.Architecture
}

func (a *AllContainerImage) GetOS() string {
	return a.OS
}

func (a *AllContainerImage) GetID() string {
	return a.ID
}

func (a *AllContainerImage) GetTarget() string {
	target := a.OS
	if a.OS == androidOS && a.Architecture != ArchMultiple {
		target += "/" + a.Architecture.String()
	}

	return target
}

func (a *AllContainerImage) GetEnv(key string) (v string, ok bool) {
	v, ok = a.Env[key]
	return
}

func (a *AllContainerImage) SetEnv(key string, value string) {
	a.Env[key] = value
}

func (a *AllContainerImage) SetMount(local string, inContainer string) {
	a.Mount[local] = inContainer
}

func (a *AllContainerImage) AppendTag(tag string) {
	a.Tags = append(a.Tags, tag)
}

// goModInit ensure a go.mod exists. If not try to generates a temporary one
func goModInit(ctx Context, image ContainerImage) error {

	goModPath := volume.JoinPathHost(ctx.WorkDirHost(), "go.mod")
	log.Infof("[i] Checking for go.mod: %s", goModPath)
	_, err := os.Stat(goModPath)
	if err == nil {
		log.Info("[✓] go.mod found")
		return nil
	}

	log.Info("[i] go.mod not found, creating a temporary one...")
	err = image.Run(ctx.Volume, Options{}, []string{"go", "mod", "init", ctx.Name})
	if err != nil {
		return fmt.Errorf("could not generate the temporary go module: %v", err)
	}

	log.Info("[✓] go.mod created")
	return nil
}

// goBuild run the go build command in the container
func goBuild(ctx Context, image ContainerImage) error {
	log.Infof("[i] Building binary...")
	// add go build command
	args := []string{"go", "build", "-trimpath"}

	// Strip debug information
	if ctx.StripDebug {
		// ensure that CGO_LDFLAGS is not overwritten as they can be passed
		// by the --env argument
		if v, ok := image.GetEnv("CGO_LDFLAGS"); ok {
			// append the ldflags after the existing CGO_LDFLAGS
			image.SetEnv("CGO_LDFLAGS", strings.Trim(v, "\"")+" -w -s")
		} else {
			// create the CGO_LDFLAGS environment variable and add the strip debug flags
			image.SetEnv("CGO_LDFLAGS", "-w -s")
		}
	}

	ldflags := ctx.LdFlags
	// honour the GOFLAGS env variable adding to existing ones
	if v, ok := image.GetEnv("GOFLAGS"); ok {
		ldflags = append(ldflags, v)
	}

	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}

	// set output folder to fyne-cross/bin/<target>
	output := volume.JoinPathContainer(ctx.Volume.BinDirContainer(), image.GetID(), ctx.Name)

	args = append(args, "-o", output)

	// enable debug mode
	if image.GetRunner().GetDebug() {
		args = append(args, "-v")
	}

	//add package
	args = append(args, ctx.Package)

	err := image.Run(ctx.Volume, Options{}, args)

	if err != nil {
		return err
	}

	log.Infof("[✓] Binary: %s", volume.JoinPathHost(ctx.BinDirHost(), image.GetID(), ctx.Name))
	return nil
}

// fynePackage package the application using the fyne cli tool
func fynePackage(ctx Context, image ContainerImage) error {
	if image.GetRunner().GetDebug() {
		err := image.Run(ctx.Volume, Options{}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
	}

	target := image.GetTarget()

	args := []string{
		fyneBin, "package",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), image.GetID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	// Enable release mode, if specified
	if ctx.Release {
		args = append(args, "-release")
	}

	// workDir default value
	workDir := ctx.WorkDirContainer()

	if image.GetOS() == androidOS {
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
	}

	// linux, darwin and freebsd targets are built by fyne-cross
	// in these cases fyne tool is used only to package the app specifying the executable flag
	if image.GetOS() == linuxOS || image.GetOS() == darwinOS || image.GetOS() == freebsdOS {
		args = append(args, "-executable", volume.JoinPathContainer(ctx.BinDirContainer(), image.GetID(), ctx.Name))
		workDir = volume.JoinPathContainer(ctx.TmpDirContainer(), image.GetID())
	}

	runOpts := Options{
		WorkDir: workDir,
	}

	err := image.Run(ctx.Volume, runOpts, args)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// fyneRelease package and release the application using the fyne cli tool
// Note: at the moment this is used only for the android builds
func fyneRelease(ctx Context, image ContainerImage) error {
	if image.GetRunner().GetDebug() {
		err := image.Run(ctx.Volume, Options{}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
	}

	target := image.GetTarget()

	args := []string{
		fyneBin, "release",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), image.GetID(), icon.Default),
		"-appBuild", ctx.AppBuild,
		"-appVersion", ctx.AppVersion,
	}

	// add appID to command, if any
	if ctx.AppID != "" {
		args = append(args, "-appID", ctx.AppID)
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("%q", strings.Join(tags, ",")))
	}

	// workDir default value
	workDir := ctx.WorkDirContainer()

	switch image.GetOS() {
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

	runOpts := Options{
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
func WindowsResource(ctx Context, image ContainerImage) (string, error) {
	args := []string{
		gowindresBin,
		"-arch", image.GetArchitecture().String(),
		"-output", ctx.Name,
		"-workdir", volume.JoinPathContainer(ctx.TmpDirContainer(), image.GetID()),
	}

	runOpts := Options{
		WorkDir: volume.JoinPathContainer(ctx.TmpDirContainer(), image.GetID()),
	}

	err := image.Run(ctx.Volume, runOpts, args)
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// copy the windows resource under the package root
	// it will be automatically linked by compiler during build
	windres := ctx.Name + ".syso"
	out := filepath.Join(ctx.Package, windres)
	err = volume.Copy(volume.JoinPathHost(ctx.TmpDirHost(), image.GetID(), windres), volume.JoinPathHost(ctx.WorkDirHost(), out))
	if err != nil {
		return "", fmt.Errorf("could not copy windows resource under the package root: %v", err)
	}

	return out, nil
}
