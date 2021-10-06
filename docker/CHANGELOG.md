# Changelog

All notable changes to the fyne-cross docker images will be documented in this file.

## fyne 1.1.x compatible

Latest versions available on Docker Hub are:

- fyneio/fyne-cross:1.1-base
- fyneio/fyne-cross:1.1-base-llvm
- fyneio/fyne-cross:1.1-base-freebsd
- fyneio/fyne-cross:1.1-android
- fyneio/fyne-cross:1.1-freebsd-amd64
- fyneio/fyne-cross:1.1-freebsd-arm64
- fyneio/fyne-cross:1.1-linux-386
- fyneio/fyne-cross:1.1-linux-arm64
- fyneio/fyne-cross:1.1-linux-arm
- fyneio/fyne-cross:1.1-windows

Release cycle won't follow the fyne-cross one, so the images will be tagged and
available on Docker Hub using the label year.month.day along with the tags
above.

Example: `fyneio/fyne-cross:1.1-base-21.03.17`

## Release 21.10.05
- Add xz-utils package to support unix packaging fyne-io/fyne#1919

## Release 21.09.29
- Update Go to v1.16.8
- Update fyne CLI to v2.1.0

### Release 21.06.07
- Update Go to v1.16.5

### Release 21.05.08
- Update Go to v1.16.4
- Update fyne CLI to v2.0.3

### Release 21.05.03
- Refactor docker images layout to ensure compatibility with previous versions of fyne-cross
- Add FreeBSD on arm64 target
- Add a dedicated docker image for macOS
- Add a dedicated docker image for Windows
- Update Go to v1.16.3
- Update fyne CLI to v2.0.2
- Update FreeBSD SDK to v12.2
- Remove the dependency from the docker/golang-cross image for the base image

> Note: the docker image for darwin is not provided anymore and need to build manually since it depends on the OSX SDK.

## fyne 1.0.x compatible

Latest versions available on Docker Hub are:
- fyneio/fyne-cross:base-latest
- fyneio/fyne-cross:darwin-latest
- fyneio/fyne-cross:linux-386-latest
- fyneio/fyne-cross:linux-arm64-latest
- fyneio/fyne-cross:linux-arm-latest
- fyneio/fyne-cross:android-latest
- fyneio/fyne-cross:freebsd-latest

Release cycle won't follow the fyne-cross one, so the images will be tagged and
available on Docker Hub using the label year.month.day along with the tags
above.

Example: `fyneio/fyne-cross:base-20.12.13`

### Release 20.12.13
- Update Go to v1.14.13
- Fix build failure for Linux mobile #19

### Release 20.12.10
- Update fyne cli to v1.4.2
> Note: this version is the last that provides Go v1.13.x

### Release 20.12.05
- Update fyne cli to v1.4.2-0.20201204171445-8f33697cf611
- Add support for Linux Wayland #10

### Release 20.11.28
- Update fyne cli to v1.4.2-0.20201127180716-f9f91c194737 fyne-io#1609

### Release 20.11.25
- Update fyne cli to v1.4.2-0.20201125075943-97ad77d2abe0 fyne-io#1538

### Release 20.11.23
- Update fyne cli to v1.4.2-0.20201122132119-67b762f56dc0 fyne-io#1527

### Release 20.11.04
- fyne cli updated to v1.4.0

# Archive

These releases occurred in the original namspace, lucor/fyne-cross

# Release 20.08.13
- Base image is based on dockercore/golang-cross@1.13.15 (Go v1.13.15)
- fyne cli updated to v1.3.3

# Release 20.07.17
- Base image is based on dockercore/golang-cross@1.13.14 (Go v1.13.14)

# Release 20.07.16
- Base image is based on dockercore/golang-cross@1.13.13 (Go v1.13.13)
- fyne cli updated to v1.3.2

# Release 20.06.07
- Base image is based on dockercore/golang-cross@1.13.12 (Go v1.13.12)
- fyne cli updated to v1.3.0

# Release 20.05.21
- Base image is based on dockercore/golang-cross@1.13.11 (Go v1.13.11)
- Android image: upgrade fyne cli tool to develop to allow build for app fyne
  develop branch

# Release 20.05.10
- Add support for FreeBSD: lucor/fyne-cross:freebsd-latest

# Release 20.05.03
- Introduce new label versioning
- Base image is based on dockercore/golang-cross@1.13.10 (Go v1.13.10)
- Add dedicated images for linux 386, arm and arm64 #25
