package command

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_makeWindowsContext(t *testing.T) {
	type args struct {
		flags *windowsFlags
	}
	tests := []struct {
		name            string
		args            args
		wantEnv         []string
		wantLdFlags     []string
		wantDockerImage string
	}{
		{
			name: "default",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						Env: &envFlag{},
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			wantEnv:         []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
			wantLdFlags:     []string{"-H windowsgui"},
			wantDockerImage: "lucor/fyne-cross:base-latest",
		},
		{
			name: "console",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						Env: &envFlag{},
					},
					TargetArch: &targetArchFlag{"386"},
					Console:    true,
				},
			},
			wantEnv:         []string{"GOOS=windows", "GOARCH=386", "CC=i686-w64-mingw32-gcc"},
			wantLdFlags:     nil,
			wantDockerImage: "lucor/fyne-cross:base-latest",
		},
		{
			name: "custom ldflags",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						Env:     &envFlag{},
						Ldflags: "-X main.version=1.2.3",
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			wantEnv:         []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
			wantLdFlags:     []string{"-X main.version=1.2.3", "-H windowsgui"},
			wantDockerImage: "lucor/fyne-cross:base-latest",
		},
		{
			name: "custom docker image",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						Env:         &envFlag{},
						DockerImage: "test",
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			wantEnv:         []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
			wantLdFlags:     []string{"-H windowsgui"},
			wantDockerImage: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetArch := []string(*tt.args.flags.TargetArch)[0]
			ctx, err := makeWindowsContext(tt.args.flags)
			if err != nil {
				t.Errorf("couldn't create Windows context: %v", err)
			}

			if len(ctx) != 1 {
				t.Errorf("len of context should be 1, but got %d", len(ctx))
			}

			if ctx[0].Architecture != Architecture(targetArch) {
				t.Errorf("architecture should be %s, but got %s", tt.args.flags.TargetArch, ctx[0].Architecture)
			}

			if ctx[0].OS != windowsOS {
				t.Errorf("architecture should be %s, but got %s", windowsOS, ctx[0].OS)
			}

			wantID := fmt.Sprintf("%s-%s", windowsOS, targetArch)
			if ctx[0].ID != wantID {
				t.Errorf("ID should be %s, but got %s", wantID, ctx[0].ID)
			}

			if !reflect.DeepEqual(ctx[0].Env, tt.wantEnv) {
				t.Errorf("expected env %s but got %s", tt.wantEnv, ctx[0].Env)
			}

			if !reflect.DeepEqual(ctx[0].LdFlags, tt.wantLdFlags) {
				t.Errorf("expected ldflags %s but got %s", tt.wantLdFlags, ctx[0].LdFlags)
			}

			if ctx[0].DockerImage != tt.wantDockerImage {
				t.Errorf("docker image should be %s, but got %s", tt.wantDockerImage, ctx[0].DockerImage)
			}
		})
	}
}
