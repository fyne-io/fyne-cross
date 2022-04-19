package command

import (
	"fmt"
	"log"
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
	engine, err := MakeEngine(dockerEngine)
	if err != nil {
		t.Skip("docker engine not found", err)
	}

	log.Println(engine.String())
	expectedCmd, err := execabs.LookPath(engine.String())
	require.NoError(t, err)
	log.Println(expectedCmd)

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
		context Context
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
				context: Context{
					Volume: vol,
					Engine: engine,
					Env:    make(map[string]string),
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, uid.Uid, uid.Gid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, dockerImage),
		},
		{
			name: "custom work dir",
			args: args{
				context: Context{
					Volume: vol,
					Engine: engine,
					Env:    make(map[string]string),
				},
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, customWorkDir, workDir, uid.Uid, uid.Gid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, customWorkDir, workDir, dockerImage),
		},
		{
			name: "cache enabled",
			args: args{
				context: Context{
					Volume:       vol,
					Engine:       engine,
					Env:          make(map[string]string),
					CacheEnabled: true,
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s -q command arg", expectedCmd, workDir, cacheDir, uid.Uid, uid.Gid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, cacheDir, dockerImage),
		},
		{
			name: "custom env variables",
			args: args{
				context: Context{
					Volume: vol,
					Engine: engine,
					Env: map[string]string{
						"GOPROXY": "proxy.example.com",
					},
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -u %s:%s --entrypoint fixuid -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s -q command arg", expectedCmd, workDir, uid.Uid, uid.Gid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewContainerEngine(tt.args.context)
			image := runner.CreateContainerImage("", "", tt.args.image)

			cmd := image.Cmd(tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if runtime.GOOS == "windows" {
				want = tt.wantWindows
			}
			if cmd != want {
				t.Errorf("Cmd()\ngot :%v\nwant:%v", cmd, want)
			}
		})
	}
}

func TestCmdEnginePodman(t *testing.T) {
	engine, err := MakeEngine(podmanEngine)
	if err != nil {
		t.Skip("podman engine not found", err)
	}

	expectedCmd, err := execabs.LookPath(engine.String())
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
		context Context
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
				context: Context{
					Volume: vol,
					Engine: engine,
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
		{
			name: "custom work dir",
			args: args{
				context: Context{
					Volume: vol,
					Engine: engine,
				},
				image: "docker.io/fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, customWorkDir, workDir, podmanFlags, dockerImage),
		},
		{
			name: "cache enabled",
			args: args{
				context: Context{
					Volume:       vol,
					Engine:       engine,
					CacheEnabled: true,
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, cacheDir, podmanFlags, dockerImage),
		},
		{
			name: "custom env variables",
			args: args{
				context: Context{
					Volume: vol,
					Engine: engine,
					Env: map[string]string{
						"GOPROXY": "proxy.example.com",
					},
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
		{
			name: "strip",
			args: args{
				context: Context{
					Volume: vol,
					Engine: engine,
					Env: map[string]string{
						"GOPROXY": "proxy.example.com",
					},
				},
				image:   "docker.io/fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewContainerEngine(tt.args.context)
			image := runner.CreateContainerImage("", "", tt.args.image)

			cmd := image.Cmd(tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if cmd != want {
				t.Errorf("Cmd()\ngot :%v\nwant:%v", cmd, want)
			}
		})
	}
}
