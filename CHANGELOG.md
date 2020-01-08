# Changelog

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
