/*
fyne-cross is a simple tool to cross compile Fyne applications

It has been inspired by xgo and uses a docker image built on top of the
golang-cross image, that includes the MinGW compiler for windows, and an OSX
SDK, along the Fyne requirements.

Supported targets are:
  -  darwin/amd64
  -  darwin/386
  -  linux/amd64
  -  linux/386
  -  linux/arm
  -  linux/arm64
  -  windows/amd64
  -  windows/386
  -  android/amd64
  -  android/386
  -  android/arm
  -  android/arm64

References
- fyne: https://fyne.io
- xgo: https://github.com/karalabe/xgo
- golang-cross: https://github.com/docker/golang-cross
- fyne-cross docker images: https://hub.docker.com/r/lucor/fyne-cross
*/
package main
