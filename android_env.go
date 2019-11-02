package main

import "fmt"

const androidPlatformAPI = 29

// Code in this file has been adapted from:
// https://github.com/golang/mobile/tree/6d0d39b2ca824d9f664f49cc75427cb806beee2e/cmd/gomobile/env.go#L322

type ndkToolchain struct {
	abi       string
	minAPI    int
	clangFlag string
}

var ndk = map[string]ndkToolchain{
	"arm": {
		abi:       "armeabi-v7a",
		minAPI:    16,
		clangFlag: fmt.Sprintf("CC=armv7a-linux-androideabi%d-clang", androidPlatformAPI),
	},
	"arm64": {
		abi:       "arm64-v8a",
		minAPI:    21,
		clangFlag: fmt.Sprintf("CC=aarch64-linux-android%d-clang", androidPlatformAPI),
	},
	"386": {
		abi:       "x86",
		minAPI:    16,
		clangFlag: fmt.Sprintf("CC=i686-linux-android%d-clang", androidPlatformAPI),
	},
	"amd64": {
		abi:       "x86_64",
		minAPI:    21,
		clangFlag: fmt.Sprintf("CC=x86_64-linux-android%d-clang", androidPlatformAPI),
	},
}
