/*
Package builder implements the build actions for the supperted OS and arch
*/
package builder

import (
	"fmt"
	"os"
	"os/user"
	"reflect"
	"testing"

	"github.com/lucor/fyne-cross/internal/volume"
)

func Test_dockerCmd(t *testing.T) {

	type args struct {
		os      string
		image   string
		vol     *volume.Volume
		env     []string
		workDir string
		command []string
		verbose bool
	}
	defaultWorkDir := "/path/to/workdir"
	vol, _ := volume.Mount(defaultWorkDir, "/path/to/cachedir")

	uid, _ := user.Current()
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "command for *nix host",
			args: args{
				image:   "lucor/fyne-cross",
				vol:     vol,
				command: []string{"echo test"},
			},
			want: fmt.Sprintf("/usr/bin/docker run --rm -t -w /app -v /path/to/workdir:/app -v /path/to/cachedir:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=%s lucor/fyne-cross echo test", uid.Uid),
		},
		{
			name: "command with custom env",
			args: args{
				image:   "lucor/fyne-cross",
				vol:     vol,
				env:     []string{"TEST=1"},
				command: []string{"echo test"},
			},
			want: fmt.Sprintf("/usr/bin/docker run --rm -t -w /app -v /path/to/workdir:/app -v /path/to/cachedir:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e TEST=1 -e fyne_uid=%s lucor/fyne-cross echo test", uid.Uid),
		},
		{
			name: "command with custom workDir",
			args: args{
				image:   "lucor/fyne-cross",
				workDir: "/path/to/workdir",
				vol:     vol,
				command: []string{"echo test"},
			},
			want: fmt.Sprintf("/usr/bin/docker run --rm -t -w /path/to/workdir -v /path/to/workdir:/app -v /path/to/cachedir:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=%s lucor/fyne-cross echo test", uid.Uid),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("GOOS", tt.args.os)
			got := dockerCmd(tt.args.image, tt.args.vol, tt.args.env, tt.args.workDir, tt.args.command, tt.args.verbose)
			if got.String() != tt.want {
				t.Errorf("dockerCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_goBuildCmd(t *testing.T) {
	type args struct {
		output string
		opts   BuildOptions
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test all params",
			args: args{
				output: "test",
				opts: BuildOptions{
					Package:    ".",
					LdFlags:    []string{"-X main.version=1.0.0"},
					Tags:       []string{"gles"},
					StripDebug: true,
					Verbose:    true,
				},
			},
			want: []string{"go", "build", "-ldflags", "'-X main.version=1.0.0 -w -s'", "-tags", "'gles'", "-o", "test", "-v", "."},
		},
		{
			name: "test no strip and verbose params",
			args: args{
				output: "test",
				opts: BuildOptions{
					Package:    ".",
					LdFlags:    []string{"-X main.version=1.0.0"},
					Tags:       []string{"gles"},
					StripDebug: false,
					Verbose:    false,
				},
			},
			want: []string{"go", "build", "-ldflags", "'-X main.version=1.0.0'", "-tags", "'gles'", "-o", "test", "."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goBuildCmd(tt.args.output, tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("goBuildCmd() = %#+v, want %#+v", got, tt.want)
			}
		})
	}
}
