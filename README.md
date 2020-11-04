# Fyne Cross

[![CI](https://github.com/fyne-io/fyne-cross/workflows/CI/badge.svg)](https://github.com/fyne-io/fyne-cross/actions?query=workflow%3ACI) [![Go Report Card](https://goreportcard.com/badge/github.com/fyne-io/fyne-cross)](https://goreportcard.com/report/github.com/fyne-io/fyne-cross) [![GoDoc](https://godoc.org/github.com/fyne-io/fyne-cross?status.svg)](http://godoc.org/github.com/fyne-io/fyne-cross) [![version](https://img.shields.io/github/v/tag/fyne-io/fyne-cross?label=version)]()

fyne-cross is a simple tool to cross compile and create distribution packages for [Fyne](https://fyne.io) applications.

It has been inspired by [xgo](https://github.com/karalabe/xgo) and uses a [docker image](https://hub.docker.com/r/fyneio/fyne-cross) built on top of the [golang-cross](https://github.com/docker/golang-cross) image, that includes the MinGW compiler for windows, and an OSX SDK, along with the Fyne requirements.

Supported targets are:
  -  darwin/amd64
  -  darwin/386
  -  freebsd/amd64
  -  linux/amd64
  -  linux/386
  -  linux/arm
  -  linux/arm64
  -  windows/amd64
  -  windows/386
  -  android
  -  ios

> Note: 
> - iOS compilation is supported only on darwin hosts. See [fyne pre-requisites](https://developer.fyne.io/started/#prerequisites) for details.
> - windows packaging for public distrubution (release mode) is supported only on windows hosts.

## Requirements

- go >= 1.13
- docker

### Installation

```
go get github.com/fyne-io/fyne-cross
```

> `fyne-cross` will be installed in GOPATH/bin, unless GOBIN is set.

### Updating docker images

To update to a newer docker image the `--pull` flag can be specified.
If set, fyne-cross will attempt to pull the image required to cross compile the application for the specified target.

For example:

```
fyne-cross linux --pull
```

will pull only the `fyne-cross:base-latest` image required to cross compile for linux target.   

## Usage

```
fyne-cross <command> [options]

The commands are:

	darwin        Build and package a fyne application for the darwin OS
	linux         Build and package a fyne application for the linux OS
	windows       Build and package a fyne application for the windows OS
	android       Build and package a fyne application for the android OS
	ios           Build and package a fyne application for the iOS OS
	freebsd       Build and package a fyne application for the freebsd OS
	version       Print the fyne-cross version information

Use "fyne-cross <command> -help" for more information about a command.
```

### Wildcards

The `arch` flag support wildcards in case want to compile against all supported GOARCH for a specified GOOS

Example:

```
fyne-cross windows -arch=*
```

is equivalent to

```
fyne-cross windows -arch=amd64,386
```

## Example

The example below cross compile and package the [fyne examples application](https://github.com/fyne-io/examples)

```
git clone https://github.com/fyne-io/examples.git
cd examples
```

### Compile and package the main example app

```
fyne-cross linux
```

> Note: by default fyne-cross will compile the package into the current dir.
>
> The command above is equivalent to: `fyne-cross linux .`

### Compile and package a particular example app

```
fyne-cross linux -output bugs ./cmd/bugs
```

## Contribute

- Fork and clone the repository
- Make and test your changes
- Open a pull request against the `develop` branch

### Contributors

See [contributors](https://github.com/fyne-io/fyne-cross/graphs/contributors) page

## Legal note

OSX/Darwin/Apple builds: 
**[Please ensure you have read and understood the Xcode license
   terms before continuing.](https://www.apple.com/legal/sla/docs/xcode.pdf)**
