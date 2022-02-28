package command

import (
	"bytes"
	"errors"
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
	// registry is the docker registry to use to pull images
	registry = "docker.io"
)

// CheckRequirements checks if the docker binary is in PATH
func CheckRequirements() error {
	_, err := execabs.LookPath("docker")
	if err == nil {
		return nil
	}
	_, err = execabs.LookPath("podman")
	if err != nil {
		return fmt.Errorf("missed requirement: docker or podman binary not found in PATH")
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

// engine returns the default engine name, if no engine is found it returns an empty string and the error.
func engine() (string, error) {
	if isEngineDocker() {
		return "docker", nil
	}

	if isEnginePodman() {
		return "podman", nil
	}

	return "", errors.New("no container engine found")
}

// isEnginePodman returns true if the engine is "podman"
func isEnginePodman() bool {
	_, err := execabs.LookPath("podman")
	return err == nil

}

// isEngineDocker return true is the engine is "docker". If "docker" is an alias to podman, so it returns false.
func isEngineDocker() bool {
	execPath, err := execabs.LookPath("docker")
	if err != nil {
		return false
	}
	// if "docker" comes from an alias (i.e. "podman-docker") should not contain the "docker" string
	out, err := execabs.Command(execPath, "--version").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "docker")
}

// Cmd returns a command to run in a new container for the specified image
func Cmd(image string, vol volume.Volume, opts Options, cmdArgs []string) *execabs.Cmd {

	// detect engine binary
	engineBinary, err := engine()
	if err != nil {
		log.Fatal(err.Error())
	}

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
	if isEnginePodman() {
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
	for _, e := range opts.Env {
		args = append(args, "-e", e)
	}

	// specify the image to use
	args = append(args, image)

	// add the command to execute
	args = append(args, cmdArgs...)

	cmd := execabs.Command(engineBinary, args...)
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
	err = Run(ctx.DockerImage, ctx.Volume, runOpts, []string{"go", "mod", "init", ctx.Name})
	if err != nil {
		return fmt.Errorf("could not generate the temporary go module: %v", err)
	}

	log.Info("[✓] go.mod created")
	return nil
}

// findGoFlags return the index of context where GOFLAGS is set, or -1 if it is not present.
func findGoFlags(ctx Context) int {
	for i, env := range ctx.Env {
		if strings.HasPrefix(env, "GOFLAGS") {
			return i
		}
	}
	return -1
}

// findCgoLdFlags return the index of context where CGO_LDFLAGS is set, or -1 if it is not present.
func findCgoLdFlags(ctx Context) int {
	for i, env := range ctx.Env {
		if strings.HasPrefix(env, "CGO_LDFLAGS") {
			return i
		}
	}
	return -1
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
		if idx := findCgoLdFlags(ctx); idx > -1 {
			// append the ldflags after the existing CGO_LDFLAGS
			ctx.Env[idx] += " -w -s"
		} else {
			// create the CGO_LDFLAGS environment variable and add the strip debug flags
			ctx.Env = append(ctx.Env, "CGO_LDFLAGS=-w -s")
		}
	}

	ldflags := ctx.LdFlags
	// add ldflags to command, if any
	if len(ldflags) > 0 {
		flags := make([]string, len(ldflags))
		for i, flag := range ldflags {
			parts := strings.Split(flag, "-") // because there could be several -X flags
			for _, p := range parts {
				if strings.HasPrefix(p, "X") {
					// We must rebuild cases
					// if the user pass "-ldflags "-X A.B=C" so we need to set it to "-X=A.B=C"
					// if the user pass "-ldflags "-X=A.B=C" so we should not change it
					sp := strings.SplitN(p, " ", 1)
					if len(sp) == 2 {
						args = append(args, "-ldflags=-X="+sp[1])
					} else {
						args = append(args, "-ldflags=-"+p)
					}
				} else {
					// others flags can be set to GOFLAGS
					flags[i] = "-ldflags=-" + p
				}
			}
		}

		// ensure that GOFLAGS is not overwritten as they can be passed
		// by the --env argument
		if idx := findGoFlags(ctx); idx > -1 {
			// append the ldflags after the existing GOFLAGS
			ctx.Env[idx] += " " + strings.Join(flags, " ")
		}
	}

	// add tags to command, if any
	tags := ctx.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, ",")))
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
		err := Run(ctx.DockerImage, ctx.Volume, Options{Debug: ctx.Debug}, []string{fyneBin, "version"})
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
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, ",")))
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
		err := Run(ctx.DockerImage, ctx.Volume, Options{Debug: ctx.Debug}, []string{fyneBin, "version"})
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
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, ",")))
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
	runner, err := engine()
	if err != nil {
		log.Fatal(err.Error())
	}
	cmd := execabs.Command(runner, "pull", registry+"/"+ctx.DockerImage)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if ctx.Debug {
		log.Debug(cmd)
	}

	err = cmd.Run()

	if ctx.Debug {
		log.Debug(buf.String())
	}

	if err != nil {
		return fmt.Errorf("could not pull the docker image: %v", err)
	}

	log.Infof("[✓] Image is up to date")
	return nil
}
