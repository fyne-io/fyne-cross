# Changelog - Fyne.io fyne-cross

## 1.6.0 - 31 Dec 2024

### Changed
- Bump github.com/stretchr/testify from 1.7.0 to 1.9.0 by @dependabot in https://github.com/fyne-io/fyne-cross/pull/233
- Ldflags where only needed with older version of zig which we have updated since then. by @Bluebugs in https://github.com/fyne-io/fyne-cross/pull/246

## 1.5.0 - 13 Apr 2024

### Changed

- Improve Docker Darwin support when using it has a host for fyne-cross (ssh-agent detection, documentation, arm64)
- Support Podman on Darwin
- Improve Android signature support
- Propagate GOFLAGS correctly
- Adjust supported Go version to match Fyne
- Update dependencies

## 1.4.0 - 13 Mar 2023

### Added

- Add support for Kubernetes
- Add ability to specify a different registry
- Support for fyne metadata

### Changed

- Pull image from fyne-cross-image repository
- Simplify `fyne-cross darwin-sdk-extract` by getting the needed files from the Apple SDK and then mounting them in the container image for each build
- Provide a darwin image and mount the SDK from the host
- Use `fyne build` for all targets

## 1.3.0 - 16 Jul 2022

### Added

- Add support for web target #92
- Add CI job to build calculator app #104
- Add support to macOS 12.x and 13.x SDKs via darwin image (osxcross) #133

### Changed

- Bump min Go version to 1.14 to align with Fyne requirements
- Update README for matching modern go command line #114

## 1.2.1 - 09 Apr 2022

### Added

- Added the `--engine` flags that allows to specify the container engine to use
  between docker and podman. The default behavior is not changed, if the flag is
  not specified fyne-cross will auto detect the engine.

### Fixed 

- Windows builds no longer pass "-H windowsgui" #97
- Multiple tags cannot be specified using the `-tags` flag #96  

## 1.2.0 - 07 Mar 2022

### Added

- Add support for FyneApp.toml #78
- Add the ability to use podman #41
- Update to use fixuid to handle mount permissions #42

## 1.1.3 - 02 Nov 2021

### Fixed

-  Building for windows fails to add icon #66
-  Fixes darwin image creation (SDK extraction) #80

## 1.1.2 - 05 Oct 2021

### Fixed

- Unsupported target operating system "linux/amd64" #74

## 1.1.1 - 29 Sep 2021

### Added

-  Support specifying target architectures for Android #52

### Changed

- Switch to x/sys/execabs for windows security fixes #57
- [base-image] update Go to v1.16.8 and Fyne CLI tool to v2.1.0 #67

## 1.1.0 - 14 May 2021

### Added

- Add darwin arm64 target #39
- Add FreeBSD on arm64 target #29
- Add the `darwin-image` command to build the darwin docker image
- Add the `local` flag for darwin to build directly from the host
- Add a dedicated docker image for macOS
- Add a dedicated docker image for Windows
- Darwin image build: add support for SDK version #45

### Changed

- Update Go to v1.16.4
- Update fyne CLI to v2.0.3
- Update FreeBSD SDK to v12.2 #29
- Refactor docker images layout to ensure compatibility with previous versions of fyne-cross

### Fixed

- Fix android keystore path is not resolved correctly
- Fix some release flags are always set even if empty
- Fix appID flag should not have a default #25
- Fix the option --env does not allow values containing comma #35

### Removed

- Remove darwin 386 target
- Remove the dependency from the docker/golang-cross image for the base image

## 1.0.0 - 13 December 2020
- Add support for "fyne release" #3
- Add support for creating packaged .tar.gz bundles on freebsd #6
- Add support for Linux Wayland #10
- Update fyne cli to v1.4.2 (fyne-io#1538 fyne-io#1527)
- Deprecate `output` flag in favour of `name`
- Fix env flag validation #14
- Fix build failure for Linux mobile #19
- Update Go to v1.14.13

## 0.9.0 - 17 October 2020
- Releaseing under project namespace with previous 2.2.1 becoming 0.9.0 in fyne-io namespace


# Archive - lucor/fyne-cross

## [2.2.1] - 2020-09-16
- Fix iOS fails with "only on darwin" when on mac #78
- Update README installation when module-aware mode is not enabled

## [2.2.0] - 2020-09-01
- Add `--pull` option to attempt to pull a newer version of the docker image #75

## [2.1.2] - 2020-08-13
- Update base image to dockercore/golang-cross@1.13.15 (Go v1.13.15)
- fyne cli updated to v1.3.3

## [2.1.1] - 2020-07-17
- Update base image to dockercore/golang-cross@1.13.14 (Go v1.13.14)

## [2.1.0] - 2020-07-16
- Add support for build flags #69
- Base image is based on dockercore/golang-cross@1.13.13 (Go v1.13.13)
- fyne cli updated to v1.3.2

## [2.0.0] - 2020-06-07
- Base image is based on dockercore/golang-cross@1.13.12 (Go v1.13.12)
- fyne cli updated to v1.3.0

## [2.0.0-beta4] - 2020-05-21
- Print fyne cli version in debug mode
- Update unit tests to work on windows
- Fix some minor linter suggestions
- Update docker base image to go v1.13.11

## [2.0.0-beta3] - 2020-05-13
- Remove package option. Package can be now specified as argument
- Fix android build when the package is not into the root dir

## [2.0.0-beta2] - 2020-05-13
- Fix build for packages not in root dir
- Fix ldflags flag not honored #62

## [2.0.0-beta1] - 2020-05-10
- Add subcommand support
- Add a flag to build as "console binary" for Windows #57
- Add support for custom env variables #59
- Add support for custom docker image #52
- Add support for FreeBSD #23

## [1.5.0] - 2020-04-13
- Add android support #37
- Add iOS support on Darwin hosts
- Issue cross compiling from Windows 10 #54
- Update to golang-cross:1.13.10 image (go v1.13.10)
- Update to fyne cli v1.2.4

## [1.4.0] - 2020-03-04
- Add ability to package with an icon using fyne/cmd #14
- Update to golang-cross:1.13.8 image (go v1.13.8) #46
- Disable android build. See #34
- Add support for passing appID to dist packaging #45
- Introduce a root folder and layout for fyne-cross output #38 
- Remove OS and Arch info from output #48
- GOCACHE folder is now mounted under $HOME/.cache/fyne-cross/go-build to cache build outputs for reuse in future builds.

## [1.3.2] - 2020-01-08
- Update to golang-cross:1.12.14 image (go v1.12.14)

## [1.3.1] - 2019-12-26
- Default binary name should be folder if none is provided [#29](https://github.com/lucor/fyne-cross/issues/29)
- Cannot build android app when not using go modules [#30](https://github.com/lucor/fyne-cross/issues/30)

## [1.3.0] - 2019-11-02
- Add Android support [#10](https://github.com/lucor/fyne-cross/issues/10)
- GOOS is not set for go get when project do not use go modules [#22](https://github.com/lucor/fyne-cross/issues/22)
- linux/386 does not work with 1.2.x [#24](https://github.com/lucor/fyne-cross/issues/24)

## [1.2.2] - 2019-10-29
- Add wildcard support for goarch [#15](https://github.com/lucor/fyne-cross/issues/15)
- Fix misleading error message when docker daemon is not available [#19](https://github.com/lucor/fyne-cross/issues/19)
- Fix build for windows/386 is failing

## [1.2.1] - 2019-10-26
- Fix fyne-cross docker image build tag

## [1.2.0] - 2019-10-26

- Fix UID is already in use [#12](https://github.com/lucor/fyne-cross/issues/12)
- Update docker image to golang-cross:1.12.12
- Add `--no-strip` flag. Since 1.1.0 by default the -w and -s flags are passed to the linker to strip binaries size omitting the symbol table, debug information and the DWARF symbol table. Specify this flag to add debug info. [#13](https://github.com/lucor/fyne-cross/issues/13)

## [1.1.0] - 2019-09-29

- Added support to `linux/arm` and `linux/arm64` targets
- Updated to golang-cross:1.12.10 image (go v1.12.10 CVE-2019-16276)

## [1.0.0] - 2019-09-07

First stable release
