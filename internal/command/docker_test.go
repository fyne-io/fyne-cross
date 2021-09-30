package command

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fyne-io/fyne-cross/internal/volume"
	"golang.org/x/sys/execabs"
)

func TestCmd(t *testing.T) {

	expectedCmd := "docker"
	if lp, err := execabs.LookPath(expectedCmd); err == nil {
		expectedCmd = lp
	}

	uid, _ := user.Current()

	workDir := filepath.Join(os.TempDir(), "fyne-cross-test", "app")
	cacheDir := filepath.Join(os.TempDir(), "fyne-cross-test", "cache")
	customWorkDir := filepath.Join(os.TempDir(), "fyne-cross-test", "custom")

	vol, err := volume.Mount(workDir, cacheDir)
	if err != nil {
		t.Fatalf("Error mounting volume test got unexpected error: %v", err)
	}

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
				image:   "fyneio/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=%s fyneio/fyne-cross command arg", expectedCmd, workDir, uid.Uid),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build fyneio/fyne-cross command arg", expectedCmd, workDir),
		},
		{
			name: "custom work dir",
			args: args{
				image: "fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: customWorkDir,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w %s -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=%s fyneio/fyne-cross command arg", expectedCmd, customWorkDir, workDir, uid.Uid),
			wantWindows: fmt.Sprintf("%s run --rm -t -w %s -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build fyneio/fyne-cross command arg", expectedCmd, customWorkDir, workDir),
		},
		{
			name: "cache enabled",
			args: args{
				image: "fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					CacheEnabled: true,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -v %s:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=%s fyneio/fyne-cross command arg", expectedCmd, workDir, cacheDir, uid.Uid),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -v %s:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build fyneio/fyne-cross command arg", expectedCmd, workDir, cacheDir),
		},
		{
			name: "custom env variables",
			args: args{
				image: "fyneio/fyne-cross",
				vol:   vol,
				opts: Options{
					Env: []string{"GOPROXY=proxy.example.com", "GOSUMDB=sum.example.com"},
				},
				cmdArgs: []string{"command", "arg"},
			},
			want:        fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com -e fyne_uid=%s fyneio/fyne-cross command arg", expectedCmd, workDir, uid.Uid),
			wantWindows: fmt.Sprintf("%s run --rm -t -w /app -v %s:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com fyneio/fyne-cross command arg", expectedCmd, workDir),
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
