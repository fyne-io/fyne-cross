package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucor/fyne-cross/v2/internal/volume"
)

func Test_makeDefaultContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	type args struct {
		flags *CommonFlags
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
				flags: &CommonFlags{},
			},
			want: Context{
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
					Env: envFlag{"TEST=true"},
				},
			},
			want: Context{
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Env:          []string{"TEST=true"},
			},
			wantErr: false,
		},
		{
			name: "custom ldflags",
			args: args{
				flags: &CommonFlags{
					Ldflags: "-X main.version=1.2.3",
				},
			},
			want: Context{
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
				flags: &CommonFlags{},
			},
			want: Context{
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
					Package: ".",
				},
			},
			want: Context{
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
					Package: "./cmd/command",
				},
			},
			want: Context{
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
					Package: volume.JoinPathHost(vol.WorkDirContainer(), "cmd/command"),
				},
			},
			want: Context{
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      volume.JoinPathHost(vol.WorkDirContainer(), "cmd/command"),
			},
			wantErr: false,
		},
		{
			name: "package absolute outside work dir",
			args: args{
				flags: &CommonFlags{
					Package: "/path/outside/workdir",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := makeDefaultContext(tt.args.flags)

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
