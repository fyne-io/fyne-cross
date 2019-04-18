# Fyne Cross

fyne-cross is a simple tool to cross compile [Fyne](https://fyne.io) applications.

It uses a docker image built on top of golang-cross (https://hub.docker.com/r/dockercore/golang-cross/),
that includes the MinGW compiler for windows, and an OSX SDK, along the Fyne requirements.

Supported targets are:
  -  windows/386
  -  darwin/amd64
  -  darwin/386
  -  linux/amd64
  -  linux/386
  -  windows/amd64

The docker image is available from https://hub.docker.com/r/lucor/fyne-cross.

## Installation

        go get github.com/lucor/fyne-cross

## Usage

        fyne-cross --targets=linux/amd64,windows/amd64,darwin/amd64 package

> Use `fyne-cross help` for more informations

## Example

The example below cross build the [fyne examples application](https://github.com/fyne-io/examples)

        git clone https://github.com/fyne-io/examples.git
        cd examples
        fyne-cross --targets=linux/amd64,windows/amd64,darwin/amd64 github.com/fyne-io/examples

Builds for the specified targets will be available under the `build` folder
