/*
Package volume implements the docker host-container volume mounting
*/
package volume

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/icon"
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

// Copy copies a resource from src to dest
func Copy(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}

// DefaultCacheDirHost returns the default cache dir on the host
func DefaultCacheDirHost() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("cannot get the path for the system cache directory on the host %s", err)
	}
	return JoinPathHost(userCacheDir, fyneCrossPrefix), nil
}

// DefaultIconHost returns the default icon path on the host
func DefaultIconHost() (string, error) {
	wd, err := DefaultWorkDirHost()
	if err != nil {
		return "", err
	}
	return JoinPathHost(wd, icon.Default), nil
}

// DefaultWorkDirHost returns the default work dir on the host
func DefaultWorkDirHost() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get the path for the work directory on the host %s", err)
	}
	return wd, nil
}

// Mount mounts the host folder into the container.
func Mount(workDirHost string, cacheDirHost string) (Volume, error) {
	var err error

	if workDirHost == "" {
		workDirHost, err = DefaultWorkDirHost()
		if err != nil {
			return Volume{}, err
		}
	}

	if cacheDirHost == "" {
		cacheDirHost, err = DefaultCacheDirHost()
		if err != nil {
			return Volume{}, err
		}
	}

	l := Volume{
		binDirHost:   JoinPathHost(workDirHost, binRelativePath),
		cacheDirHost: cacheDirHost,
		distDirHost:  JoinPathHost(workDirHost, distRelativePath),
		tmpDirHost:   JoinPathHost(workDirHost, tmpRelativePath),
		workDirHost:  workDirHost,
	}

	err = createHostDirs(l)
	if err != nil {
		return l, err
	}

	return l, nil
}

// JoinPathContainer joins any number of path elements into a single path,
// separating them with the Container OS specific Separator.
func JoinPathContainer(elem ...string) string {
	return path.Clean(strings.Join(elem, "/"))
}

// JoinPathHost joins any number of path elements into a single path,
// separating them with the Host OS specific Separator.
func JoinPathHost(elem ...string) string {
	return filepath.Clean(filepath.Join(elem...))
}

// Zip compress the source file into a zip archive
func Zip(source string, archive string) error {

	sourceData, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("could not read the source file content: %s", err)
	}
	defer sourceData.Close()

	// Get the file information
	sourceInfo, err := sourceData.Stat()
	if err != nil {
		return fmt.Errorf("could not get the source file info: %s", err)
	}

	// Create the archive file
	archiveFile, err := os.Create(archive)
	if err != nil {
		return fmt.Errorf("could not create the zip archive file: %s", err)
	}

	// Create a new zip archive.
	zipWriter := zip.NewWriter(archiveFile)

	header, err := zip.FileInfoHeader(sourceInfo)
	if err != nil {
		return fmt.Errorf("could not create the file header: %s", err)
	}
	header.Method = zip.Deflate

	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not add the source file to the zip archive: %s", err)
	}

	_, err = io.Copy(w, sourceData)
	if err != nil {
		return fmt.Errorf("could not write the source file content into the zip archive: %s", err)
	}

	// Make sure to check the error on Close.
	err = zipWriter.Close()
	if err != nil {
		return fmt.Errorf("could not close the zip archive: %s", err)
	}

	return archiveFile.Close()
}

// Volume represents the fyne-cross projec layout
type Volume struct {
	binDirHost   string
	cacheDirHost string
	distDirHost  string
	tmpDirHost   string
	workDirHost  string
}

// BinDirContainer returns the fyne-cross bin dir on the container
func (l Volume) BinDirContainer() string {
	return binDirContainer
}

// BinDirHost returns the fyne-cross bin dir on the host
func (l Volume) BinDirHost() string {
	return l.binDirHost
}

// CacheDirContainer returns the fyne-cross cache dir on the container
func (l Volume) CacheDirContainer() string {
	return cacheDirContainer
}

// CacheDirHost returns the fyne-cross cache dir on the host
func (l Volume) CacheDirHost() string {
	return l.cacheDirHost
}

// DistDirContainer returns the fyne-cross distribution dir on the container
func (l Volume) DistDirContainer() string {
	return distDirContainer
}

// DistDirHost returns the fyne-cross distribution dir on the host
func (l Volume) DistDirHost() string {
	return l.distDirHost
}

// GoCacheDirContainer returns the fyne-cross Go cache dir on the container
func (l Volume) GoCacheDirContainer() string {
	return goCacheDirContainer
}

// TmpDirContainer returns the fyne-cross temporary dir on the container
func (l Volume) TmpDirContainer() string {
	return tmpDirContainer
}

// TmpDirHost returns the fyne-cross temporary dir on the host
func (l Volume) TmpDirHost() string {
	return l.tmpDirHost
}

// WorkDirContainer returns the working dir on the host
func (l Volume) WorkDirContainer() string {
	return workDirContainer
}

// WorkDirHost returns the working dir on the host
func (l Volume) WorkDirHost() string {
	return l.workDirHost
}

// createHostDirs creates the fyne-cross dirs on the host, if not exists
func createHostDirs(l Volume) error {
	dirs := []string{
		l.binDirHost,
		l.cacheDirHost,
		l.distDirHost,
		l.tmpDirHost,
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create the fyne-cross directory %s: %s", dir, err)
		}
	}
	return nil
}
