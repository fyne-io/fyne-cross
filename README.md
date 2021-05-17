# Fyne Cross

[![CI](https://github.com/fyne-io/fyne-cross/workflows/CI/badge.svg)](https://github.com/fyne-io/fyne-cross/actions?query=workflow%3ACI) [![Go Report Card](https://goreportcard.com/badge/github.com/fyne-io/fyne-cross)](https://goreportcard.com/report/github.com/fyne-io/fyne-cross) [![GoDoc](https://godoc.org/github.com/fyne-io/fyne-cross?status.svg)](http://godoc.org/github.com/fyne-io/fyne-cross) [![version](https://img.shields.io/github/v/tag/fyne-io/fyne-cross?label=version)]()

fyne-cross is a simple tool to cross compile and create distribution packages
for [Fyne](https://fyne.io) applications using docker images that include Linux,
the MinGW compiler for Windows, FreeBSD, and a macOS SDK, along with the Fyne
requirements.

Supported targets are:
  -  darwin/amd64
  -  darwin/arm64
  -  freebsd/amd64
  -  freebsd/arm64
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
> - macOS packaging for public distrubution (release mode) is supported only on darwin hosts.
> - windows packaging for public distrubution (release mode) is supported only on windows hosts.
> - starting from v1.1.0:
>   - cross-compile from NOT `darwin` (i.e. linux) to `darwin`: the image with the macOS SDK is no more available via docker hub and has to be built manually, see the [Build the darwin image](#build_darwin_image) section below.
>   - cross-compile from `darwin` to `darwin` by default will use under the hood the fyne CLI tool and requires Go and the macOS SDK installed on the host.

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

## <a name="build_darwin_image"></a>Build the docker image for OSX/Darwin/Apple cross-compiling
The docker image for darwin is not provided via docker hub and need to build manually since it depends on the macOS SDK.

**[Please ensure you have read and understood the Xcode license
   terms before continuing.](https://www.apple.com/legal/sla/docs/xcode.pdf)**

To build the image:
1. [Download Command Line Tools for Xcode](https://developer.apple.com/download/more) >= 12.4 (macOS SDK 11.x)
2. Run: `fyne-cross darwin-image --xcode-path /path/to/Command_Line_Tools_for_Xcode_12.5.dmg`

The command above will:
- install the dependencies required by [osxcross](https://github.com/tpoechtrager/osxcross) to package the macOS SDK and compile the macOS cross toolchain.
- package the macOS SDK
- compile the macOS cross toolchain
- build the `fyneio/fyne-cross:<ver>-darwin` image that will be used by fyne-cross

> NOTE: the creation of the image may take several minutes and may require more than 25 GB of free disk space.

### [EXPERIMENTAL] Build using a different SDK version

By default fyne-cross will attempt to auto-detect the latest version of SDK provided by the Xcode package. If for any reason a different SDK version is required, it can be specified using the `--sdk-version` flag.

Example:

`fyne-cross darwin-image --sdk-version 11.1 --xcode-path /path/to/Command_Line_Tools_for_Xcode_12.4.dmg`

> Note: this feature is marked as EXPERIMENTAL

## Contribute

- Fork and clone the repository
- Make and test your changes
- Open a pull request against the `develop` branch

### Contributors

See [contributors](https://github.com/fyne-io/fyne-cross/graphs/contributors) page

## Credits

- [osxcross](https://github.com/tpoechtrager/osxcross) for the macOS Cross toolchain for Linux
- [golang-cross](https://github.com/docker/golang-cross) for the inspiration and the docker images used in the initial versions
- [xgo](https://github.com/karalabe/xgo) for the inspiration
