// Run a command line helper for various Fyne tools.
package main

import (
	"os"
	"os/user"
	"reflect"
	"strings"
	"testing"
)

func Test_parseTargets(t *testing.T) {
	type args struct {
		targetList string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Target list cannot be empty",
			args:    args{targetList: ""},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Invalid target",
			args:    args{targetList: "invalid/amd64"},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Invalid target 2",
			args:    args{targetList: "linux/amd64,invalid/amd64"},
			want:    []string{"linux/amd64"},
			wantErr: true,
		},
		{
			name:    "Invalid target 3",
			args:    args{targetList: "linux/*amd64"},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Valid target",
			args:    args{targetList: "linux/amd64"},
			want:    []string{"linux/amd64"},
			wantErr: false,
		},
		{
			name:    "Valid targets trim space",
			args:    args{targetList: "linux/amd64, darwin/386"},
			want:    []string{"linux/amd64", "darwin/386"},
			wantErr: false,
		},
		{
			name:    "Valid wildcard targets",
			args:    args{targetList: "linux/*"},
			want:    []string{"linux/amd64", "linux/386", "linux/arm", "linux/arm64"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTargets(tt.args.targetList)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerBuilder_targetOutput(t *testing.T) {
	type fields struct {
		output string
		pkg    string
		target string
	}

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "default *nix plaform",
			fields: fields{
				output: "fyne-example",
				pkg:    "fyne-io/fyne-example",
				target: "linux/amd64",
			},
			want: "fyne-example-linux-amd64",
		},
		{
			name: "default windows plaform",
			fields: fields{
				output: "fyne-example",
				pkg:    "fyne-io/fyne-example",
				target: "windows/386",
			},
			want: "fyne-example-windows-386.exe",
		},
		{
			name: "custom output *nix plaform",
			fields: fields{
				output: "test",
				pkg:    "fyne-io/fyne-example",
				target: "linux/amd64",
			},
			want: "test-linux-amd64",
		},
		{
			name: "custom output windows plaform",
			fields: fields{
				output: "test",
				pkg:    "fyne-io/fyne-example",
				target: "windows/386",
			},
			want: "test-windows-386.exe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osAndArch := strings.Split(tt.fields.target, "/")
			d := &dockerBuilder{
				output: tt.fields.output,
				pkg:    tt.fields.pkg,
				target: tt.fields.target,
				os:     osAndArch[0],
				arch:   osAndArch[1],
			}
			got, err := d.targetOutput()
			if (err != nil) != tt.wantErr {
				t.Errorf("dockerBuilder.targetOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerBuilder.targetOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerBuilder_verbosityFlag(t *testing.T) {
	type fields struct {
		verbose bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "verbosity enabled",
			fields: fields{
				verbose: true,
			},
			want: "-v",
		},
		{
			name: "verbosity disabled",
			fields: fields{
				verbose: false,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dockerBuilder{
				verbose: tt.fields.verbose,
			}
			if got := d.verbosityFlag(); got != tt.want {
				t.Errorf("dockerBuilder.verbosityFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerBuilder_defaultArgs(t *testing.T) {
	// current work dir
	wd, _ := os.Getwd()

	// cache dir
	cd, _ := os.UserCacheDir()

	// current user id
	u, _ := user.Current()
	uid := u.Uid

	type fields struct {
		pkg     string
		workDir string
		goPath  string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "default workdir",
			fields: fields{
				pkg:     "fyne-io/fyne-example",
				workDir: wd,
				goPath:  cd,
			},
			want: []string{
				"run", "--rm", "-t",
				"-w", "/app",
				"-v", wd + ":/app",
				"-v", cd + ":/go",
				"-e", "fyne_uid=" + uid,
			},
		},
		{
			name: "custom workdir",
			fields: fields{
				pkg:     "fyne-io/fyne-example",
				workDir: "/home/fyne",
				goPath:  "/tmp/cache",
			},
			want: []string{
				"run", "--rm", "-t",
				"-w", "/app",
				"-v", "/home/fyne:/app",
				"-v", "/tmp/cache:/go",
				"-e", "fyne_uid=" + uid,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dockerBuilder{
				pkg:     tt.fields.pkg,
				workDir: tt.fields.workDir,
				goPath:  tt.fields.goPath,
			}
			if got := d.defaultArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerBuilder.defaultArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerBuilder_goBuildArgs(t *testing.T) {
	type fields struct {
		targets    []string
		output     string
		pkg        string
		workDir    string
		verbose    bool
		ldflags    string
		stripDebug bool
		target     string
	}

	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "verbosity enabled, linux",
			fields: fields{
				verbose:    true,
				pkg:        "fyne-io/fyne-example",
				workDir:    "/code/test",
				output:     "test",
				stripDebug: true,
				target:     "linux/amd64",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-ldflags", "'-w -s'",
				"-o", "build/test-linux-amd64",
				"-a",
				"-v",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "verbosity disabled, windows/amd64",
			fields: fields{
				verbose: false,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				ldflags: "-X main.version=1.0.0",
				target:  "windows/amd64",
			},

			want: []string{
				dockerImage,
				"go", "build",
				"-ldflags", "'-H windowsgui -X main.version=1.0.0'",
				"-o", "build/test-windows-amd64.exe",
				"-a",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "verbosity disabled, windows/386",
			fields: fields{
				verbose: false,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				ldflags: "-X main.version=1.0.0",
				target:  "windows/386",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-ldflags", "'-H windowsgui -X main.version=1.0.0'",
				"-o", "build/test-windows-386.exe",
				"-a",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "default settings from current dir darwin",
			fields: fields{
				output: "fyne-example",
				pkg:    "fyne-io/fyne-example",
				target: "darwin/amd64",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-o", "build/fyne-example-darwin-amd64",
				"-a",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "ldflags, linux",
			fields: fields{
				verbose:    true,
				pkg:        "fyne-io/fyne-example",
				workDir:    "/code/test",
				output:     "test",
				ldflags:    "-X main.version=1.0.0",
				stripDebug: true,
				target:     "linux/amd64",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-ldflags", "'-w -s -X main.version=1.0.0'",
				"-o", "build/test-linux-amd64",
				"-a",
				"-v",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "arm",
			fields: fields{
				verbose: true,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				target:  "linux/arm",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-tags", "'gles'",
				"-o", "build/test-linux-arm",
				"-a",
				"-v",
				"fyne-io/fyne-example",
			},
		},
		{
			name: "arm64",
			fields: fields{
				verbose: true,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				target:  "linux/arm64",
			},
			want: []string{
				dockerImage,
				"go", "build",
				"-tags", "'gles'",
				"-o", "build/test-linux-arm64",
				"-a",
				"-v",
				"fyne-io/fyne-example",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osAndArch := strings.Split(tt.fields.target, "/")
			d := &dockerBuilder{
				target:     tt.fields.target,
				output:     tt.fields.output,
				pkg:        tt.fields.pkg,
				workDir:    tt.fields.workDir,
				verbose:    tt.fields.verbose,
				ldflags:    tt.fields.ldflags,
				stripDebug: tt.fields.stripDebug,
				os:         osAndArch[0],
				arch:       osAndArch[1],
			}
			got, err := d.goBuildArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("dockerBuilder.goBuildArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerBuilder.goBuildArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerBuilder_goEnvArgs(t *testing.T) {
	type fields struct {
		targets    []string
		output     string
		pkg        string
		workDir    string
		verbose    bool
		ldflags    string
		stripDebug bool
		target     string
	}

	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "verbosity enabled, linux",
			fields: fields{
				verbose:    true,
				pkg:        "fyne-io/fyne-example",
				workDir:    "/code/test",
				output:     "test",
				stripDebug: true,
				target:     "linux/amd64",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=linux", "-e", "GOARCH=amd64", "-e", "CC=gcc",
			},
		},
		{
			name: "verbosity disabled, windows/amd64",
			fields: fields{
				verbose: false,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				ldflags: "-X main.version=1.0.0",
				target:  "windows/amd64",
			},

			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=windows", "-e", "GOARCH=amd64", "-e", "CC=x86_64-w64-mingw32-gcc",
			},
		},
		{
			name: "verbosity disabled, windows/386",
			fields: fields{
				verbose: false,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				ldflags: "-X main.version=1.0.0",
				target:  "windows/386",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=windows", "-e", "GOARCH=386", "-e", "CC=i686-w64-mingw32-gcc",
			},
		},
		{
			name: "default settings from current dir darwin",
			fields: fields{
				pkg:    "fyne-io/fyne-example",
				target: "darwin/amd64",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=darwin", "-e", "GOARCH=amd64", "-e", "CC=o32-clang",
			},
		},
		{
			name: "ldflags, linux",
			fields: fields{
				verbose:    true,
				pkg:        "fyne-io/fyne-example",
				workDir:    "/code/test",
				output:     "test",
				ldflags:    "-X main.version=1.0.0",
				stripDebug: true,
				target:     "linux/amd64",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=linux", "-e", "GOARCH=amd64", "-e", "CC=gcc",
			},
		},
		{
			name: "arm",
			fields: fields{
				verbose: true,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				target:  "linux/arm",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=linux", "-e", "GOARCH=arm", "-e", "CC=arm-linux-gnueabihf-gcc", "-e", "GOARM=7",
			},
		},
		{
			name: "arm64",
			fields: fields{
				verbose: true,
				pkg:     "fyne-io/fyne-example",
				workDir: "/code/test",
				output:  "test",
				target:  "linux/arm64",
			},
			want: []string{
				"-e", "CGO_ENABLED=1",
				"-e", "GOOS=linux", "-e", "GOARCH=arm64", "-e", "CC=aarch64-linux-gnu-gcc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osAndArch := strings.Split(tt.fields.target, "/")
			d := &dockerBuilder{
				target:     tt.fields.target,
				output:     tt.fields.output,
				pkg:        tt.fields.pkg,
				workDir:    tt.fields.workDir,
				verbose:    tt.fields.verbose,
				ldflags:    tt.fields.ldflags,
				stripDebug: tt.fields.stripDebug,
				os:         osAndArch[0],
				arch:       osAndArch[1],
			}
			got := d.goEnvArgs()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerBuilder.goEnvArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
