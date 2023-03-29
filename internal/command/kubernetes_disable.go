//go:build !k8s
// +build !k8s

package command

import (
	"errors"
	"flag"
)

var errNotImplemented error = errors.New("kubernetes support not built in. Compile fyne-cross with `-tag k8s` to enable it")

func kubernetesFlagSet(_ *flag.FlagSet, _ *CommonFlags) {
}

func checkKubernetesClient() (err error) {
	return errNotImplemented
}

func newKubernetesContainerRunner(context Context) (containerEngine, error) {
	return nil, errNotImplemented
}
