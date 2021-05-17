package command

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fyne-io/fyne-cross/internal/volume"
)

func Test_makeDefaultContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	type args struct {
		flags *CommonFlags
		args  []string
	}
	tests := []struct {
		name    string
		args    args
		want    Context
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
			},
			wantErr: false,
		},
		{
			name: "custom env",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Env:      envFlag{"TEST=true"},
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Env:          []string{"TEST=true"},
			},
			wantErr: false,
		},
		{
			name: "custom env containg =",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Env:      envFlag{"GOFLAGS=-mod=vendor"},
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Env:          []string{"GOFLAGS=-mod=vendor"},
			},
			wantErr: false,
		},
		{
			name: "custom ldflags",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Ldflags:  "-X main.version=1.2.3",
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				LdFlags:      []string{"-X main.version=1.2.3"},
			},
			wantErr: false,
		},
		{
			name: "package default",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
			},
			wantErr: false,
		},
		{
			name: "package dot",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
				args: []string{"."},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
			},
			wantErr: false,
		},
		{
			name: "package relative",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
				args: []string{"./cmd/command"},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      "./cmd/command",
			},
			wantErr: false,
		},
		{
			name: "package absolute",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
				args: []string{volume.JoinPathHost(vol.WorkDirHost(), "cmd/command")},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      "./cmd/command",
			},
			wantErr: false,
		},
		{
			name: "package absolute outside work dir",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
				},
				args: []string{os.TempDir()},
			},
			wantErr: true,
		},
		{
			name: "custom tags",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Tags:     tagsFlag{"hints", "gles"},
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Tags:         []string{"hints", "gles"},
			},
			wantErr: false,
		},
		{
			name: "invalid app build",
			args: args{
				flags: &CommonFlags{
					AppBuild: 0,
				},
			},
			want:    Context{},
			wantErr: true,
		},
		{
			name: "release mode enabled",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Release:  true,
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Release:      true,
			},
			wantErr: false,
		},
		{
			name: "app version",
			args: args{
				flags: &CommonFlags{
					AppBuild:   1,
					AppVersion: "1.0",
					Release:    true,
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				AppVersion:   "1.0",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Release:      true,
			},
			wantErr: false,
		},
		{
			name: "deprecate output flag in favour of name",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Name:     "./test",
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Name:         "./test",
			},
			wantErr: true,
		},
		{
			name: "valid name",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					Name:     "test",
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Name:         "test",
			},
			wantErr: false,
		},
		{
			name: "appID",
			args: args{
				flags: &CommonFlags{
					AppBuild: 1,
					AppID:    "com.example.test",
					Name:     "test",
				},
			},
			want: Context{
				AppBuild:     "1",
				AppID:        "com.example.test",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Name:         "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := makeDefaultContext(tt.args.flags, tt.args.args)

			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want, ctx)
		})
	}
}

func mockDefaultVolume() (volume.Volume, error) {
	rootDir, err := volume.DefaultWorkDirHost()
	if err != nil {
		return volume.Volume{}, err
	}
	cacheDir, err := volume.DefaultCacheDirHost()
	if err != nil {
		return volume.Volume{}, err
	}
	return volume.Mount(rootDir, cacheDir)
}
