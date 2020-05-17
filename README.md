# Fyne Cross

[![CI](https://github.com/lucor/fyne-cross/workflows/CI/badge.svg)](https://github.com/lucor/fyne-cross/actions?query=workflow%3ACI) [![Go Report Card](https://goreportcard.com/badge/github.com/lucor/fyne-cross)](https://goreportcard.com/report/github.com/lucor/fyne-cross) [![GoDoc](https://godoc.org/github.com/lucor/fyne-cross?status.svg)](http://godoc.org/github.com/lucor/fyne-cross) [![GitHub tag](https://img.shields.io/github/tag/lucor/fyne-cross.svg)]()

fyne-cross is a simple tool to cross compile and create distribution packages for [Fyne](https://fyne.io) applications.

It has been inspired by [xgo](https://github.com/karalabe/xgo) and uses a [docker image](https://hub.docker.com/r/lucor/fyne-cross) built on top of the [golang-cross](https://github.com/docker/golang-cross) image, that includes the MinGW compiler for windows, and an OSX SDK, along with the Fyne requirements.

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

> Note: iOS compilation is supported only on darwin hosts. See [fyne README mobile](https://github.com/fyne-io/fyne/blob/v1.2.4/README-mobile.md#ios) for pre-requisites.

## Requirements

- go >= 1.13
- docker

### Installation

```
go get github.com/lucor/fyne-cross/v2/cmd/fyne-cross
```

### Development release

To install a preview of the v2 version or help in testing:

```
go get github.com/lucor/fyne-cross/v2/cmd/fyne-cross@develop
```

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

See [contributors](https://github.com/lucor/fyne-cross/graphs/contributors) page

## Legal note

OSX/Darwin/Apple builds: 
**[Please ensure you have read and understood the Xcode license
   terms before continuing.](https://www.apple.com/legal/sla/docs/xcode.pdf)**
