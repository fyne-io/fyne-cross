package command

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fyne-io/fyne-cross/internal/volume"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/execabs"
)

func TestCmdEngineDocker(t *testing.T) {
	engineBinary, err := engine()
	if err != nil {
		t.Skip("engine not found", err)
	}

	if isEnginePodman() {
		t.Skip("engine found: podman")
	}

	expectedCmd, err := execabs.LookPath(engineBinary)
	require.NoError(t, err)

	uid, _ := user.Current()

	workDir := filepath.Join(os.TempDir(), "fyne-cross-test", "app")
	cacheDir := filepath.Join(os.TempDir(), "fyne-cross-test", "cache")
	customWorkDir := filepath.Join(os.TempDir(), "fyne-cross-test", "custom")

	vol, err := volume.Mount(workDir, cacheDir)
	if err != nil {
		t.Fatalf("Error mounting volume test got unexpected error: %v", err)
	}

	dockerImage := "docker.io/fyneio/fyne-cross"

	type args struct {
		image   string
		vol     volume.Volume
		opts    Options
		cmdArgs []string
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantWindows string
	}{
		{
			name: "default",
			args: args{
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, uid.Uid, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, dockerImage),
		},
		{
			name: "custom work dir",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, customWorkDir, workDir, uid.Uid, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, customWorkDir, workDir, dockerImage),
		},
		{
			name: "cache enabled",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					CacheEnabled: true,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, cacheDir, uid.Uid, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, cacheDir, dockerImage),
		},
		{
			name: "custom env variables",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					Env: []string{"GOPROXY=proxy.example.com", "GOSUMDB=sum.example.com"},
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com %s -q command arg", expectedCmd, workDir, uid.Uid, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com %s command arg", expectedCmd, workDir, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Cmd(tt.args.image, tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if runtime.GOOS == "windows" {
				want = tt.wantWindows
			}
			if cmd != want {
				t.Errorf("Cmd() command = %v, want %v", cmd, want)
			}
		})
	}
}

func TestCmdEnginePodman(t *testing.T) {
	engineBinary, err := engine()
	if err != nil {
		t.Skip("engine not found", err)
	}

	if isEngineDocker() {
		t.Skip("engine found: docker")
	}

	expectedCmd, err := execabs.LookPath(engineBinary)
	require.NoError(t, err)

	workDir := filepath.Join(os.TempDir(), "fyne-cross-test", "app")
	cacheDir := filepath.Join(os.TempDir(), "fyne-cross-test", "cache")
	customWorkDir := filepath.Join(os.TempDir(), "fyne-cross-test", "custom")

	vol, err := volume.Mount(workDir, cacheDir)
	if err != nil {
		t.Fatalf("Error mounting volume test got unexpected error: %v", err)
	}

	dockerImage := "docker.io/fyneio/fyne-cross"
	podmanFlags := "--userns keep-id -e use_podman=1"

	type args struct {
		image   string
		vol     volume.Volume
		opts    Options
		cmdArgs []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
		{
			name: "custom work dir",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, customWorkDir, workDir, podmanFlags, dockerImage),
		},
		{
			name: "cache enabled",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					CacheEnabled: true,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -v %s:/go:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, podmanFlags, cacheDir, dockerImage),
		},
		{
			name: "custom env variables",
			args: args{
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					Env: []string{"GOPROXY=proxy.example.com", "GOSUMDB=sum.example.com"},
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com %s -q command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Cmd(tt.args.image, tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if cmd != want {
				t.Errorf("Cmd() command = %v, want %v", cmd, want)
			}
		})
	}
}
