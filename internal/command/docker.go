package command

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
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

// Options define the options for the docker run command
type Options struct {
	CacheEnabled bool              // CacheEnabled if true enable go build cache
	Engine       Engine            // Engine is the container engine to use
	Env          map[string]string // Env is the list of custom env variable to set
	WorkDir      string            // WorkDir set the workdir, default to volume's workdir
	Debug        bool              // Debug if true enable log verbosity
}

// Cmd returns a command to run in a new container for the specified image
func Cmd(image string, vol volume.Volume, opts Options, cmdArgs []string) *execabs.Cmd {

	// define workdir
	w := vol.WorkDirContainer()
	if opts.WorkDir != "" {
		w = opts.WorkDir
	}

	args := []string{
		"run", "--rm", "-t",
		"-w", w, // set workdir
		"-v", fmt.Sprintf("%s:%s:z", vol.WorkDirHost(), vol.WorkDirContainer()), // mount the working dir
	}

	// mount the cache dir if cache is enabled
	if opts.CacheEnabled {
		args = append(args, "-v", fmt.Sprintf("%s:%s:z", vol.CacheDirHost(), vol.CacheDirContainer()))
	}

	// handle settings related to engine
	if opts.Engine.IsPodman() {
		args = append(args, "--userns", "keep-id", "-e", "use_podman=1")
	} else {
		// docker: pass current user id to handle mount permissions on linux and MacOS
		if runtime.GOOS != "windows" {
			u, err := user.Current()
			if err == nil {
				args = append(args, "-u", fmt.Sprintf("%s:%s", u.Uid, u.Gid))
				args = append(args, "--entrypoint", "fixuid")
				if !opts.Debug {
					// silent fixuid if not debug mode
					cmdArgs = append([]string{"-q"}, cmdArgs...)
				}
			}
		}
	}

	// add default env variables
	args = append(args,
		"-e", "CGO_ENABLED=1", // enable CGO
		"-e", fmt.Sprintf("GOCACHE=%s", vol.GoCacheDirContainer()), // mount GOCACHE to reuse cache between builds
	)

	// add custom env variables
	for k, v := range opts.Env {
		env := k + "=" + v
		if strings.Contains(v, "=") {
			// engine requires to double quote the env var when value contains
			// the `=` char
			env = fmt.Sprintf("%q", env)
		}
		args = append(args, "-e", env)
	}

	// specify the image to use
	args = append(args, image)

	// add the command to execute
	args = append(args, cmdArgs...)

	cmd := execabs.Command(opts.Engine.Binary, args...)
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
	runOpts := Options{Debug: ctx.Debug, Engine: ctx.Engine}
	err = Run(ctx.DockerImage, ctx.Volume, runOpts, []string{"go", "mod", "init", ctx.Name})
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
	args := []string{"go", "build", "-trimpath"}

	// Strip debug information
	if ctx.StripDebug {
		// ensure that CGO_LDFLAGS is not overwritten as they can be passed
		// by the --env argument
		if v, ok := ctx.Env["CGO_LDFLAGS"]; ok {
			// append the ldflags after the existing CGO_LDFLAGS
			ctx.Env["CGO_LDFLAGS"] = strings.Trim(v, "\"") + " -w -s"
		} else {
			// create the CGO_LDFLAGS environment variable and add the strip debug flags
			ctx.Env["CGO_LDFLAGS"] = "-w -s"
		}
	}

	ldflags := ctx.LdFlags
	// honour the GOFLAGS env variable adding to existing ones
	if v, ok := ctx.Env["GOFLAGS"]; ok {
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
	output := volume.JoinPathContainer(ctx.Volume.BinDirContainer(), ctx.ID, ctx.Name)

	args = append(args, "-o", output)

	// enable debug mode
	if ctx.Debug {
		args = append(args, "-v")
	}

	//add package
	args = append(args, ctx.Package)
	runOpts := Options{
		CacheEnabled: ctx.CacheEnabled,
		Engine:       ctx.Engine,
		Env:          ctx.Env,
		Debug:        ctx.Debug,
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)

	if err != nil {
		return err
	}

	log.Infof("[✓] Binary: %s", volume.JoinPathHost(ctx.BinDirHost(), ctx.ID, ctx.Name))
	return nil
}

// fynePackage package the application using the fyne cli tool
func fynePackage(ctx Context) error {

	if ctx.Debug {
		err := Run(ctx.DockerImage, ctx.Volume, Options{Debug: ctx.Debug, Engine: ctx.Engine}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
	}

	target := ctx.OS
	if ctx.OS == androidOS && ctx.Architecture != ArchMultiple {
		target += "/" + ctx.Architecture.String()
	}

	args := []string{
		fyneBin, "package",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID, icon.Default),
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

	if ctx.OS == androidOS {
		workDir = volume.JoinPathContainer(workDir, ctx.Package)
	}

	// linux, darwin and freebsd targets are built by fyne-cross
	// in these cases fyne tool is used only to package the app specifying the executable flag
	if ctx.OS == linuxOS || ctx.OS == darwinOS || ctx.OS == freebsdOS {
		args = append(args, "-executable", volume.JoinPathContainer(ctx.BinDirContainer(), ctx.ID, ctx.Name))
		workDir = volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID)
	}

	runOpts := Options{
		CacheEnabled: ctx.CacheEnabled,
		WorkDir:      workDir,
		Debug:        ctx.Debug,
		Engine:       ctx.Engine,
		Env:          ctx.Env,
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)
	if err != nil {
		return fmt.Errorf("could not package the Fyne app: %v", err)
	}
	return nil
}

// fyneRelease package and release the application using the fyne cli tool
// Note: at the moment this is used only for the android builds
func fyneRelease(ctx Context) error {

	if ctx.Debug {
		err := Run(ctx.DockerImage, ctx.Volume, Options{Debug: ctx.Debug, Engine: ctx.Engine}, []string{fyneBin, "version"})
		if err != nil {
			return fmt.Errorf("could not get fyne cli %s version: %v", fyneBin, err)
		}
	}

	target := ctx.OS
	if ctx.OS == androidOS && ctx.Architecture != ArchMultiple {
		target += "/" + ctx.Architecture.String()
	}

	args := []string{
		fyneBin, "release",
		"-os", target,
		"-name", ctx.Name,
		"-icon", volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID, icon.Default),
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

	switch ctx.OS {
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
		CacheEnabled: ctx.CacheEnabled,
		WorkDir:      workDir,
		Debug:        ctx.Debug,
		Engine:       ctx.Engine,
		Env:          ctx.Env,
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

	args := []string{
		gowindresBin,
		"-arch", ctx.Architecture.String(),
		"-output", ctx.Name,
		"-workdir", volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID),
	}

	runOpts := Options{
		Debug:   ctx.Debug,
		Engine:  ctx.Engine,
		WorkDir: volume.JoinPathContainer(ctx.TmpDirContainer(), ctx.ID),
	}

	err := Run(ctx.DockerImage, ctx.Volume, runOpts, args)
	if err != nil {
		return "", fmt.Errorf("could not package the Fyne app: %v", err)
	}

	// copy the windows resource under the package root
	// it will be automatically linked by compiler during build
	windres := ctx.Name + ".syso"
	out := filepath.Join(ctx.Package, windres)
	err = volume.Copy(volume.JoinPathHost(ctx.TmpDirHost(), ctx.ID, windres), volume.JoinPathHost(ctx.WorkDirHost(), out))
	if err != nil {
		return "", fmt.Errorf("could not copy windows resource under the package root: %v", err)
	}

	return out, nil
}

// pullImage attempts to pull a newer version of the docker image
func pullImage(ctx Context) error {
	if !ctx.Pull {
		return nil
	}

	log.Infof("[i] Checking for a newer version of the docker image: %s", ctx.DockerImage)

	buf := bytes.Buffer{}

	// run the command inside the container
	cmd := execabs.Command(ctx.Engine.Binary, "pull", ctx.DockerImage)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if ctx.Debug {
		log.Debug(cmd)
	}

	err := cmd.Run()

	if ctx.Debug {
		log.Debug(buf.String())
	}

	if err != nil {
		return fmt.Errorf("could not pull the docker image: %v", err)
	}

	log.Infof("[✓] Image is up to date")
	return nil
}
