/*
Package volume implements the docker host-container volume mounting
*/
package volume

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	fyneCrossPrefix = "fyne-cross"

	binRelativePath  = "fyne-cross/bin"
	distRelativePath = "fyne-cross/dist"
	tmpRelativePath  = "fyne-cross/tmp"

	binDirContainer     = "/app/" + binRelativePath
	cacheDirContainer   = "/go"
	goCacheDirContainer = "/go/go-build"
	distDirContainer    = "/app/" + distRelativePath
	tmpDirContainer     = "/app/" + tmpRelativePath
	workDirContainer    = "/app"
)

// Volume represents the fyne-cross projec layout
type Volume struct {
	binDirHost   string
	cacheDirHost string
	distDirHost  string
	tmpDirHost   string
	workDirHost  string
}

// Mount mounts the host folder into the container.
func Mount(workDirHost string, cacheDirHost string) (*Volume, error) {
	var err error

	if workDirHost == "" {
		workDirHost, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Cannot get the path for the work directory on the host %s", err)
		}
	}

	if cacheDirHost == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("Cannot get the path for the system cache directory on the host %s", err)
		}
		cacheDirHost = JoinPathHost(userCacheDir, fyneCrossPrefix)

	}

	l := &Volume{
		binDirHost:   JoinPathHost(workDirHost, binRelativePath),
		cacheDirHost: cacheDirHost,
		distDirHost:  JoinPathHost(workDirHost, distRelativePath),
		tmpDirHost:   JoinPathHost(workDirHost, tmpRelativePath),
		workDirHost:  workDirHost,
	}

	return l, nil
}

// CreateHostDirs creates the fyne-cross dirs on the host, if not exists
func (l *Volume) CreateHostDirs() error {
	dirs := []string{
		l.binDirHost,
		l.cacheDirHost,
		l.distDirHost,
		l.tmpDirHost,
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("Cannot create the fyne-cross directory %s: %s", dir, err)
		}
	}
	return nil
}

// BinDirHost returns the fyne-cross bin dir on the host
func (l *Volume) BinDirHost() string {
	return l.binDirHost
}

// CacheDirHost returns the fyne-cross cache dir on the host
func (l *Volume) CacheDirHost() string {
	return l.cacheDirHost
}

// DistDirHost returns the fyne-cross distribution dir on the host
func (l *Volume) DistDirHost() string {
	return l.distDirHost
}

// TmpDirHost returns the fyne-cross temporary dir on the host
func (l *Volume) TmpDirHost() string {
	return l.tmpDirHost
}

// WorkDirHost returns the working dir on the host
func (l *Volume) WorkDirHost() string {
	return l.workDirHost
}

// BinDirContainer returns the fyne-cross bin dir on the container
func (l *Volume) BinDirContainer() string {
	return binDirContainer
}

// CacheDirContainer returns the fyne-cross cache dir on the container
func (l *Volume) CacheDirContainer() string {
	return cacheDirContainer
}

// GoCacheDirContainer returns the fyne-cross Go cache dir on the container
func (l *Volume) GoCacheDirContainer() string {
	return goCacheDirContainer
}

// DistDirContainer returns the fyne-cross distribution dir on the container
func (l *Volume) DistDirContainer() string {
	return distDirContainer
}

// TmpDirContainer returns the fyne-cross temporary dir on the container
func (l *Volume) TmpDirContainer() string {
	return tmpDirContainer
}

// WorkDirContainer returns the working dir on the host
func (l *Volume) WorkDirContainer() string {
	return workDirContainer
}

// JoinPathContainer joins any number of path elements into a single path,
// separating them with the Container OS specific Separator.
func JoinPathContainer(elem ...string) string {
	return strings.Join(elem, "/")
}

// JoinPathHost joins any number of path elements into a single path,
// separating them with the Host OS specific Separator.
func JoinPathHost(elem ...string) string {
	return filepath.Join(elem...)
}
