package command

import (
	"runtime"
	"testing"

	"github.com/lucor/fyne-cross/v2/internal/volume"
)

func TestCmd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TODO update for windows")
	}

	vol, err := volume.Mount("/tmp/fyne-cross-test/app", "/tmp/fyne-cross-test/cache")
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
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				image:   "lucor/fyne-cross",
				vol:     vol,
				opts:    Options{},
				cmdArgs: []string{"command", "arg"},
			},
			want: "/usr/bin/docker run --rm -t -w /app -v /tmp/fyne-cross-test/app:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=1000 lucor/fyne-cross command arg",
		},
		{
			name: "custom work dir",
			args: args{
				image: "lucor/fyne-cross",
				vol:   vol,
				opts: Options{
					WorkDir: "/tmp/fyne-cross-test/custom-wd",
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: "/usr/bin/docker run --rm -t -w /tmp/fyne-cross-test/custom-wd -v /tmp/fyne-cross-test/app:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=1000 lucor/fyne-cross command arg",
		},
		{
			name: "cache enabled",
			args: args{
				image: "lucor/fyne-cross",
				vol:   vol,
				opts: Options{
					CacheEnabled: true,
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: "/usr/bin/docker run --rm -t -w /app -v /tmp/fyne-cross-test/app:/app -v /tmp/fyne-cross-test/cache:/go -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e fyne_uid=1000 lucor/fyne-cross command arg",
		},
		{
			name: "custom env variables",
			args: args{
				image: "lucor/fyne-cross",
				vol:   vol,
				opts: Options{
					Env: []string{"GOPROXY=proxy.example.com", "GOSUMDB=sum.example.com"},
				},
				cmdArgs: []string{"command", "arg"},
			},
			want: "/usr/bin/docker run --rm -t -w /app -v /tmp/fyne-cross-test/app:/app -e CGO_ENABLED=1 -e GOCACHE=/go/go-build -e GOPROXY=proxy.example.com -e GOSUMDB=sum.example.com -e fyne_uid=1000 lucor/fyne-cross command arg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Cmd(tt.args.image, tt.args.vol, tt.args.opts, tt.args.cmdArgs).String()
			if cmd != tt.want {
				t.Errorf("Cmd() command = %v, want %v", cmd, tt.want)
			}
		})
	}
}
