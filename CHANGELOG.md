# Changelog - Fyne.io fyne-cross

## [1.0.0]
- Add support for "fyne release" #3
- Add support for creating packaged .tar.gz bundles on freebsd #6
- Add support for Linux Wayland #10
- Update fyne cli to v1.4.2 (fyne-io#1538 fyne-io#1527)
- Deprecate `output` flag in favour of `name`
- Fix env flag validation #14
- Fix build failure for Linux mobile #19
- Update Go to v1.14.13

## [0.9.0]
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
