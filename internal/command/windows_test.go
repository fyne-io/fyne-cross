package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_makeWindowsContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	type args struct {
		flags *windowsFlags
		args  []string
	}
	tests := []struct {
		name    string
		args    args
		want    []Context
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					ID:           "windows-amd64",
					OS:           "windows",
					Architecture: "amd64",
					Env:          []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
					LdFlags:      []string{"-H=windowsgui"},
					DockerImage:  windowsImage,
				},
			},
		},
		{
			name: "console",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
					},
					TargetArch: &targetArchFlag{"386"},
					Console:    true,
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					ID:           "windows-386",
					OS:           "windows",
					Architecture: "386",
					Env:          []string{"GOOS=windows", "GOARCH=386", "CC=i686-w64-mingw32-gcc"},
					DockerImage:  windowsImage,
				},
			},
		},
		{
			name: "custom ldflags",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
						Ldflags:  "-X main.version=1.2.3",
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					ID:           "windows-amd64",
					OS:           "windows",
					Architecture: "amd64",
					Env:          []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
					LdFlags:      []string{"-X main.version=1.2.3", "-H windowsgui"},
					DockerImage:  windowsImage,
				},
			},
		},
		{
			name: "custom docker image",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						AppBuild:    1,
						DockerImage: "test",
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					ID:           "windows-amd64",
					OS:           "windows",
					Architecture: "amd64",
					Env:          []string{"GOOS=windows", "GOARCH=amd64", "CC=x86_64-w64-mingw32-gcc"},
					LdFlags:      []string{"-H=windowsgui"},
					DockerImage:  "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				ctx, err := makeWindowsContext(tt.args.flags, tt.args.args)
				if tt.wantErr {
					require.NotNil(t, err)
					return
				}
				require.Nil(t, err)
				assert.Equal(t, tt.want, ctx)
			})
		})
	}
}
