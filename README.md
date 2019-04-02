# Fyne Cross Compile Docker Image

This repo contains a Dockerfile for building an image which is used to cross compile [Fyne](https://fyne.io) applications. 

Built on top of [golang-cross](https://hub.docker.com/r/dockercore/golang-cross/), it includes the MinGW compiler for windows, and an OSX SDK, along the Fyne requirements.

This image is available from https://hub.docker.com/lucor/fyne-cross.

## Usage

Assuming the [fyne app](https://fyne.io/develop/) is located under: `$GOPATH/fyne-example`

Cross compiling build can be done using the commands below:

### linux

    docker run --rm -ti -v $GOPATH:/go -w /go/src/fyne-example \
        -e CGO_ENABLED=1 -e GOOS=linux -e CC=gcc \
        lucor/fyne-cross \
        bash -c "go get -v ./...; go build -o fyne-example-linux"

### osx

    docker run --rm -ti -v $GOPATH:/go -w /go/src/fyne-example \
        -e CGO_ENABLED=1 -e GOOS=darwin -e CC=o32-clang \
        lucor/fyne-cross \
        bash -c "go get -v ./...; go build -o fyne-example-osx"

### windows

    docker run --rm -ti -v $GOPATH:/go -w /go/src/fyne-example \
        -e CGO_ENABLED=1 -e GOOS=windows -e CC=x86_64-w64-mingw32-gcc \
        lucor/fyne-cross \
        bash -c "go get -v ./...; go build -o fyne-example-windows.exe"
