package command

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"

	"golang.org/x/sys/execabs"
)

// Options define the options for the docker run command
type options struct {
	WorkDir string // WorkDir set the workdir, default to volume's workdir
}

type localContainerEngine struct {
	baseEngine

	engine *Engine

	pull         bool
	cacheEnabled bool
}

func newLocalContainerEngine(context Context) (containerEngine, error) {
	return &localContainerEngine{
		baseEngine: baseEngine{
			env:  context.Env,
			tags: context.Tags,
			vol:  context.Volume,
		},
		engine:       &context.Engine,
		pull:         context.Pull,
		cacheEnabled: context.CacheEnabled,
	}, nil
}

type localContainerImage struct {
	baseContainerImage

	runner *localContainerEngine
}

var _ containerEngine = (*localContainerEngine)(nil)
var _ closer = (*localContainerImage)(nil)

func (r *localContainerEngine) createContainerImage(arch Architecture, OS string, image string) containerImage {
	ret := r.createContainerImageInternal(arch, OS, image, func(base baseContainerImage) containerImage {
		return &localContainerImage{
			baseContainerImage: base,
			runner:             r,
		}
	})

	// mount the cache dir if cache is enabled
	if r.cacheEnabled {
		ret.SetMount("cache", r.vol.CacheDirHost(), r.vol.CacheDirContainer())
	}

	return ret
}

func (*localContainerImage) close() error {
	return nil
}

func AppendEnv(args []string, environs map[string]string, quoteNeeded bool) []string {
	for k, v := range environs {
		env := k + "=" + v
		if quoteNeeded && strings.Contains(v, "=") {
			// engine requires to double quote the env var when value contains
			// the `=` char
			env = fmt.Sprintf("%q", env)
		}
		args = append(args, "-e", env)
	}
	return args
}

func (i *localContainerImage) Engine() containerEngine {
	return i.runner
}

// Cmd returns a command to run in a new container for the specified image
func (i *localContainerImage) cmd(vol volume.Volume, opts options, cmdArgs []string) *execabs.Cmd {

	// define workdir
	w := vol.WorkDirContainer()
	if opts.WorkDir != "" {
		w = opts.WorkDir
	}

	args := []string{
		"run", "--rm", "-t",
		"-w", w, // set workdir
	}

	for _, mountPoint := range i.mount {
		args = append(args, "-v", fmt.Sprintf("%s:%s:z", mountPoint.localHost, mountPoint.inContainer))
	}

	// handle settings related to engine
	if i.runner.engine.IsPodman() {
		args = append(args, "--userns", "keep-id", "-e", "use_podman=1")
	} else {
		// docker: pass current user id to handle mount permissions on linux and MacOS
		if runtime.GOOS != "windows" {
			u, err := user.Current()
			if err == nil {
				args = append(args, "-u", fmt.Sprintf("%s:%s", u.Uid, u.Gid))
				args = append(args, "--entrypoint", "fixuid")
				if !debugging() {
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
	args = AppendEnv(args, i.runner.env, i.env["GOOS"] != freebsdOS)
	args = AppendEnv(args, i.env, i.env["GOOS"] != freebsdOS)

	// specify the image to use
	args = append(args, i.DockerImage)

	// add the command to execute
	args = append(args, cmdArgs...)

	cmd := execabs.Command(i.runner.engine.Binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// Run runs a command in a new container for the specified image
func (i *localContainerImage) Run(vol volume.Volume, opts options, cmdArgs []string) error {
	cmd := i.cmd(vol, opts, cmdArgs)
	log.Debug(cmd)
	return cmd.Run()
}

// pullImage attempts to pull a newer version of the docker image
func (i *localContainerImage) Prepare() error {
	if !i.runner.pull {
		return nil
	}

	log.Infof("[i] Checking for a newer version of the docker image: %s", i.DockerImage)

	buf := bytes.Buffer{}

	// run the command inside the container
	cmd := execabs.Command(i.runner.engine.Binary, "pull", i.DockerImage)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	log.Debug(cmd)

	err := cmd.Run()

	log.Debug(buf.String())

	if err != nil {
		return fmt.Errorf("could not pull the docker image: %v", err)
	}

	log.Infof("[✓] Image is up to date")
	return nil
}

func (i *localContainerImage) Finalize(packageName string) error {
	// move the dist package into the "dist" folder
	srcPath := volume.JoinPathHost(i.runner.vol.TmpDirHost(), i.ID(), packageName)
	distFile := volume.JoinPathHost(i.runner.vol.DistDirHost(), i.ID(), packageName)
	err := os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		return fmt.Errorf("could not create the dist package dir: %v", err)
	}

	err = os.Rename(srcPath, distFile)
	if err != nil {
		return err
	}

	log.Infof("[✓] Package: %q", distFile)

	return nil
}
