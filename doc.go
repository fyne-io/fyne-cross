/*
fyne-cross is a simple tool to cross compile Fyne applications (https://fyne.io)

It uses a docker image built on top of golang-cross (https://hub.docker.com/r/dockercore/golang-cross/),
that includes the MinGW compiler for windows, and an OSX SDK, along the Fyne requirements.

Supported targets are:
  -  windows/386
  -  darwin/amd64
  -  darwin/386
  -  linux/amd64
  -  linux/386
  -  windows/amd64

*/
package main
