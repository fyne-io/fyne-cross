# Fyne Cross

[![CircleCI](https://circleci.com/gh/lucor/fyne-cross.svg?style=svg)](https://circleci.com/gh/lucor/fyne-cross) [![Go Report Card](https://goreportcard.com/badge/github.com/lucor/fyne-cross)](https://goreportcard.com/report/github.com/lucor/fyne-cross) [![GoDoc](https://godoc.org/github.com/lucor/fyne-cross?status.svg)](http://godoc.org/github.com/lucor/fyne-cross) [![GitHub tag](https://img.shields.io/github/tag/lucor/fyne-cross.svg)]()

fyne-cross is a simple tool to cross compile and create distribution packages for [Fyne](https://fyne.io) applications.

It has been inspired by [xgo](https://github.com/karalabe/xgo) and uses a [docker image](https://hub.docker.com/r/lucor/fyne-cross) built on top of the [golang-cross](https://github.com/docker/golang-cross) image,
that includes the MinGW compiler for windows, and an OSX SDK, along with the Fyne requirements.

Supported targets are:
  -  darwin/amd64
  -  darwin/386
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

- go
- docker

## Installation

        go get github.com/lucor/fyne-cross/cmd/fyne-cross

### Development release

To install a preview of the next version or help in testing:

        go get github.com/lucor/fyne-cross/cmd/fyne-cross@develop

## Usage

        fyne-cross --targets=linux/amd64,windows/amd64,darwin/amd64 package

> Use `fyne-cross help` for more informations

### Wildcards

The `targets` flag support wildcards in case want to compile against all supported GOARCH for a specified GOOS

Example:

        fyne-cross --targets=linux/*

is equivalent to

       fyne-cross --targets=linux/amd64,linux/386,linux/arm64,linux/arm

## Example

The example below cross build the [fyne examples application](https://github.com/fyne-io/examples)

        git clone https://github.com/fyne-io/examples.git
        cd examples
        fyne-cross --targets=linux/amd64,windows/amd64,darwin/amd64 github.com/fyne-io/examples

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
