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
	"github.com/stretchr/testify/assert"
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
	mountFlag := ":z"
	if runtime.GOOS == darwinOS && runtime.GOARCH == string(ArchArm64) {
		// When running on darwin with a Arm64, we rely on going through a VM setup that doesn't allow the :z
		mountFlag = ""
	}

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
		opts    options
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
				opts:    options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app%s --platform linux/%s --user %s -e HOME=/tmp -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, mountFlag, runtime.GOARCH, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z --platform linux/amd64 -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, dockerImage),
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
				opts: options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w %s -v %s:/app%s --platform linux/%s --user %s -e HOME=/tmp -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, customWorkDir, workDir, mountFlag, runtime.GOARCH, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app:z --platform linux/amd64 -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, customWorkDir, workDir, dockerImage),
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
				opts:    options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app%s -v %s:/go%s --platform linux/%s --user %s -e HOME=/tmp -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, mountFlag, cacheDir, mountFlag, runtime.GOARCH, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z -v %s:/go:z --platform linux/amd64 -e CGO_ENABLED=1 -e GOCACHE=/go/go-build %s command arg", expectedCmd, workDir, cacheDir, dockerImage),
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
				opts:    options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app%s --platform linux/%s --user %s -e HOME=/tmp -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, mountFlag, runtime.GOARCH, uid.Uid, dockerImage),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z --platform linux/amd64 -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := newContainerEngine(tt.args.context)
			assert.Nil(t, err)
			image := runner.createContainerImage("", "", tt.args.image)

			cmd := image.(*localContainerImage).cmd(tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if runtime.GOOS == "windows" {
				want = tt.wantWindows
			}
			if cmd != want {
				t.Errorf("cmd()\ngot :%v\nwant:%v", cmd, want)
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
	podmanFlags := "--userns keep-id -e use_podman=1 --arch=amd64"

	type args struct {
		context Context
		image   string
		vol     volume.Volume
		opts    options
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
				opts:    options{},
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
				opts: options{
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
				opts:    options{},
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
				opts:    options{},
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
				opts:    options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app:z %s -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com %s command arg", expectedCmd, workDir, podmanFlags, dockerImage),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := newContainerEngine(tt.args.context)
			assert.Nil(t, err)
			image := runner.createContainerImage("", "", tt.args.image)

			cmd := image.(*localContainerImage).cmd(tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			want := tt.want
			if cmd != want {
				t.Errorf("cmd()\ngot :%v\nwant:%v", cmd, want)
			}
		})
	}
}

func TestAppendEnv(t *testing.T) {
	type args struct {
		args        []string
		env         map[string]string
		quoteNeeded bool
	}
	tests := []struct {
		name      string
		args      args
		wantStart []string
		wantEnd   [][2]string
	}{
		{
			name: "empty",
			args: args{
				args:        []string{},
				env:         map[string]string{},
				quoteNeeded: true,
			},
			wantStart: []string{},
			wantEnd:   [][2]string{},
		},
		{
			name: "quote needed",
			args: args{
				args:        []string{},
				env:         map[string]string{"VAR": "value"},
				quoteNeeded: true,
			},
			wantStart: []string{},
			wantEnd:   [][2]string{{"-e", "VAR=value"}},
		},
		{
			name: "quote not needed",
			args: args{
				args:        []string{},
				env:         map[string]string{"VAR": "value"},
				quoteNeeded: false,
			},
			wantStart: []string{},
			wantEnd:   [][2]string{{"-e", "VAR=value"}},
		},
		{
			name: "multiple",
			args: args{
				args:        []string{},
				env:         map[string]string{"VAR": "value", "VAR2": "value2"},
				quoteNeeded: true,
			},
			wantStart: []string{},
			wantEnd:   [][2]string{{"-e", "VAR=value"}, {"-e", "VAR2=value2"}},
		},
		{
			name: "multiple with args",
			args: args{
				args:        []string{"arg1", "arg2"},
				env:         map[string]string{"VAR": "value", "VAR2": "value2"},
				quoteNeeded: true,
			},
			wantStart: []string{"arg1", "arg2"},
			wantEnd:   [][2]string{{"-e", "VAR=value"}, {"-e", "VAR2=value2"}},
		},
		{
			name: "multiple with args and equal sign require quoting values",
			args: args{
				args:        []string{"arg1", "arg2"},
				env:         map[string]string{"VAR": "value", "VAR2": "value2=2"},
				quoteNeeded: true,
			},
			wantStart: []string{"arg1", "arg2"},
			wantEnd:   [][2]string{{"-e", "VAR=value"}, {"-e", "VAR2=\"value2=2\""}},
		},
		{
			name: "multiple with args and equal sign do not require quoting values",
			args: args{
				args:        []string{"arg1", "arg2"},
				env:         map[string]string{"VAR": "value", "VAR2": "value2=2"},
				quoteNeeded: false,
			},
			wantStart: []string{"arg1", "arg2"},
			wantEnd:   [][2]string{{"-e", "VAR=value"}, {"-e", "VAR2=value2=2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendEnv(tt.args.args, tt.args.env, tt.args.quoteNeeded)
			var i int
			for _, v := range tt.wantStart {
				assert.Equal(t, v, got[i])
				i++
			}
			for ; i < len(got); i += 2 {
				found := false
				for k, v := range tt.wantEnd {
					if v[0] == got[i] && v[1] == got[i+1] {
						tt.wantEnd = append(tt.wantEnd[:k], tt.wantEnd[k+1:]...)
						found = true
						break
					}
				}
				assert.Equal(t, true, found)
			}
			assert.Equal(t, 0, len(tt.wantEnd))
		})
	}
}

func TestMain(m *testing.M) {
	os.Unsetenv("SSH_AUTH_SOCK")
	os.Exit(m.Run())
}
