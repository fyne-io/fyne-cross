package command

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/lucor/fyne-cross/v2/internal/icon"
	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// fyneBin is the path of the fyne binary into the docker image
	fyneBin = "/usr/local/bin/fyne"
	// gowindresBin is the path of the gowindres binary into the docker image
	gowindresBin = "/usr/local/bin/gowindres"
)

// CheckRequirements checks if the docker binary is in PATH
func CheckRequirements() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("missed requirement: docker binary not found in PATH")
	}
	return nil
}

// Options define the options for the docker run command
type Options struct {
	CacheEnabled bool     // CacheEnabled if true enable go build cache
	Env          []string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	WorkDir      string   // WorkDir set the workdir, default to volume's workdir
	Debug        bool     // Debug if true enable log verbosity
}

// Cmd returns a command to run in a new container for the specified image
func Cmd(image string, vol volume.Volume, opts Options, cmdArgs []string) *exec.Cmd {

	// define workdir
	w := vol.WorkDirContainer()
	if opts.WorkDir != "" {
		w = opts.WorkDir
	}

	args := []string{
		"run", "--rm", "-t",
		"-w", w, // set workdir
		"-v", fmt.Sprintf("%s:%s", vol.WorkDirHost(), vol.WorkDirContainer()), // mount the working dir
	}

	// mount the cache dir if cache is enabled
	if opts.CacheEnabled {
		args = append(args, "-v", fmt.Sprintf("%s:%s", vol.CacheDirHost(), vol.CacheDirContainer()))
	}

	// add default env variables
	args = append(args,
		"-e", "CGO_ENABLED=1", // enable CGO
		"-e", fmt.Sprintf("GOCACHE=%s", vol.GoCacheDirContainer()), // mount GOCACHE to reuse cache between builds
	)

	// add custom env variables
	for _, e := range opts.Env {
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
	args = append(args, cmdArgs...)

	// run the command inside the container
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// Run runs a command in a new container for the specified image
func Run(image string, vol volume.Volume, opts Options, cmdArgs []string) error {
	cmd := Cmd(image, vol, opts, cmdArgs)
	if opts.Debug {
		log.Debug(cmd)
	}
	return cmd.Run()
}

// goModInit ensure a go.mod exists. If not try to generates a temporary one
func goModInit(ctx Context) error {

	goModPath := volume.JoinPathHost(ctx.WorkDirHost(), "go.mod")
	log.Infof("[i] Checking for go.mod: %s", goModPath)
	_, err := os.Stat(goModPath)
	if err == nil {
		log.Info("[✓] go.mod found")
		return nil
	}

	log.Info("[i] go.mod not found, creating a temporary one...")
	runOpts := Options{Debug: ctx.Debug}
	err = Run(ctx.DockerImage, ctx.Volume, runOpts, []string{"go", "mod", "init", ctx.Output})
	if err != nil {
		return fmt.Errorf("could not generate the temporary go module: %v", err)
	}

	log.Info("[✓] go.mod created")
	return nil
}

// goBuild run the go build command in the container
func goBuild(ctx Context) error {
	log.Infof("[i] Building binary...")
	// add go build command
	args := []string{"go", "build"}

	ldflags := ctx.LdFlags
	// Strip debug information
	if ctx.StripDebug {
		ldflags = append(ldflags, "-w", "-s")
	}

	// add ldflags to command, if any
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", fmt.Sprintf("'%s'", strings.Join(ldflags, " ")))
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")))
	}

	// set output folder to fyne-cross/bin/<target>
	output := volume.JoinPathContainer(ctx.Volume.BinDirContainer(), ctx.ID, ctx.Output)

	args = append(args, "-o", output)

	// enable debug mode
	if ctx.Debug {
		args = append(args, "-v")
	}

	//add package
	args = append(args, ctx.Package)
	runOpts := Options{
		CacheEnabled: ctx.CacheEnabled,
		Env:          ctx.Env,
		Debug:        ctx.Debug,
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)

	if err != nil {
		return err
	}

	log.Infof("[✓] Binary: %s", volume.JoinPathHost(ctx.BinDirHost(), ctx.ID, ctx.Output))
	return nil
}

// fynePackage package the application using the fyne cli tool
func fynePackage(ctx Context) error {

	args := []string{
		fyneBin, "package",
		"-os", ctx.OS,
		"-name", ctx.Output,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID, icon.Default),
		"-appID", ctx.AppID,
	}

	// workDir default value
	workDir := ctx.WorkDirContainer()

	if ctx.OS == androidOS || ctx.OS == iosOS {
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
	}

	// set executable flag for linux and darwin targets
	if ctx.OS == linuxOS || ctx.OS == darwinOS {
		args = append(args, "-executable", volume.JoinPathContainer(ctx.BinDirContainer(), ctx.ID, ctx.Output))
		workDir = volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID)
	}

	runOpts := Options{
		CacheEnabled: ctx.CacheEnabled,
		WorkDir:      workDir,
		Debug:        ctx.Debug,
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// WindowsResource create a windows resource under the project root
// that will be automatically linked by compliler during the build
func WindowsResource(ctx Context) (string, error) {

	windres := ctx.Output + ".syso"

	args := []string{
		gowindresBin,
		"-arch", ctx.Architecture.String(),
		"-output", ctx.Output,
		"-workdir", volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID),
	}

	runOpts := Options{
		Debug:   ctx.Debug,
		WorkDir: volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID),
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)
	if err != nil {
		return windres, fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// copy the windows resource under the project root
	// it will be automatically linked by compiler during build
	err = volume.Copy(volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, windres), volume.JoinPathHost(ctx.WorkDirHost(), windres))
	if err != nil {
		return windres, fmt.Errorf("could not copy windows resource under the project root: %v", err)
	}

	return windres, nil
}
