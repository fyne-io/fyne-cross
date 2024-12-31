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
  -  windows/arm64
  -  windows/386
  -  android ([multiple architectures](https://developer.android.com/ndk/guides/abis))
  -  android/386
  -  android/amd64
  -  android/arm
  -  android/arm64
  -  ios

> Note: 
> - iOS compilation is supported only on darwin hosts. See [fyne pre-requisites](https://developer.fyne.io/started/#prerequisites) for details.
> - macOS packaging for public distribution (release mode) is supported only on darwin hosts.
> - windows packaging for public distribution (release mode) is supported only on windows hosts.
> - starting from v1.1.0:
>   - cross-compile from NOT `darwin` (i.e. linux) to `darwin`: requires a copy of the macOS SDK on the host. The fyne-cross `darwin-sdk-extractor` command can be used to extract the SDK from the XCode CLI Tool file, see the [Extract the macOS SDK](#extract_macos_sdk) section below.
>   - cross-compile from `darwin` to `darwin` by default will use under the hood the fyne CLI tool and requires Go and the macOS SDK installed on the host.
> - starting from v1.4.0, Arm64 hosts are supported for all platforms except Android.

## Requirements

- go >= 1.14
- docker

### Installation

For go >= 1.16:
```
go install github.com/fyne-io/fyne-cross@latest
```

To install a fyne-cross with kubernetes engine support:
```
go install -tags k8s github.com/fyne-io/fyne-cross@latest
```

For older go:
```
GO111MODULE=on go get -u github.com/fyne-io/fyne-cross
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

## <a name="extract_macos_sdk"></a>Extract the macOS SDK for OSX/Darwin/Apple cross-compiling
cross-compile from NOT `darwin` (i.e. linux) to `darwin` requires a copy of the macOS SDK on the host. 
The fyne-cross `darwin-sdk-extractor` command can be used to extract the SDK from the XCode CLI Tool file.

**[Please ensure you have read and understood the Xcode license
   terms before continuing.](https://www.apple.com/legal/sla/docs/xcode.pdf)**

To extract the SDKs:
1. [Download Command Line Tools for Xcode](https://developer.apple.com/download/all/?q=Command%20Line%20Tools) 12.4 (macOS SDK 11.x)
2. Run: `fyne-cross darwin-sdk-extract --xcode-path /path/to/Command_Line_Tools_for_Xcode_12.5.dmg`
  * Once extraction has been done, you should have a SDKs directory created. This directory contains at least 2 SDKs (ex. `SDKs/MacOSX12.3.sdk/` and `SDKs/MacOSX13.3.sdk/` in Command_Line_Tools_for_Xcode_14.3.1.dmg)
3. Specify explicitly which SDK you want to use in your fyne-cross command with --macosx-sdk-path: `fyne-cross darwin --macosx-sdk-path /full/path/to/SDKs/MacOSX12.3.sdk -app-id your.app.id`

> Note: current version supports only MacOS SDK 11.3

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
