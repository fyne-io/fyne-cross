package command

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/execabs"
)

const (
	autodetectEngine = ""
	dockerEngine     = "docker"
	podmanEngine     = "podman"
)

type Engine struct {
	Name   string
	Binary string
}

func (e Engine) String() string {
	return e.Name
}

func (e Engine) IsDocker() bool {
	return e.Name == dockerEngine
}

func (e Engine) IsPodman() bool {
	return e.Name == podmanEngine
}

// MakeEngine returns a new container engine. Pass empty string to autodetect
func MakeEngine(e string) (Engine, error) {
	switch e {
	case dockerEngine:
		binaryPath, err := execabs.LookPath(dockerEngine)
		if err != nil {
			return Engine{}, fmt.Errorf("docker binary not found in PATH")
		}
		return Engine{Name: dockerEngine, Binary: binaryPath}, nil
	case podmanEngine:
		binaryPath, err := execabs.LookPath(podmanEngine)
		if err != nil {
			return Engine{}, fmt.Errorf("podman binary not found in PATH")
		}
		return Engine{Name: podmanEngine, Binary: binaryPath}, nil
	case "":
		binaryPath, err := execabs.LookPath(dockerEngine)
		if err != nil {
			// check for podman engine
			binaryPath, err := execabs.LookPath(podmanEngine)
			if err != nil {
				return Engine{}, fmt.Errorf("engine binary not found in PATH")
			}
			return Engine{Name: podmanEngine, Binary: binaryPath}, nil
		}
		// docker binary found, check if it is an alias to podman
		// if "docker" comes from an alias (i.e. "podman-docker") should not contain the "docker" string
		out, err := execabs.Command(binaryPath, "--version").Output()
		if err != nil {
			return Engine{}, fmt.Errorf("could not detect engine version: %s", out)
		}
		lout := strings.ToLower(string(out))
		switch {
		case strings.Contains(lout, dockerEngine):
			return Engine{Name: dockerEngine, Binary: binaryPath}, nil
		case strings.Contains(lout, podmanEngine):
			return Engine{Name: podmanEngine, Binary: binaryPath}, nil
		default:
			return Engine{}, fmt.Errorf("could not detect engine version: %s", out)
		}
	default:
		return Engine{}, errors.New("unsupported container engine")
	}
}
