package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_makeAndroidContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	type args struct {
		flags *androidFlags
		args  []string
	}
	tests := []struct {
		name    string
		args    args
		want    []Context
		wantErr bool
	}{
		{
			name: "keystore path must be relative to project root",
			args: args{
				flags: &androidFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
						AppID:    "com.example.test",
					},
					Keystore:   "/tmp/my.keystore",
					TargetArch: &targetArchFlag{string(ArchMultiple)},
				},
			},
			want:    []Context{},
			wantErr: true,
		},
		{
			name: "keystore path does not exist",
			args: args{
				flags: &androidFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
						AppID:    "com.example.test",
					},
					Keystore:   "my.keystore",
					TargetArch: &targetArchFlag{string(ArchMultiple)},
				},
			},
			want:    []Context{},
			wantErr: true,
		},
		{
			name: "default",
			args: args{
				flags: &androidFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
						AppID:    "com.example.test",
					},
					Keystore:   "testdata/my.keystore",
					TargetArch: &targetArchFlag{string(ArchMultiple)},
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					AppID:        "com.example.test",
					ID:           androidOS,
					OS:           androidOS,
					Architecture: ArchMultiple,
					DockerImage:  androidImage,
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					Keystore:     "/app/testdata/my.keystore",
				},
			},
			wantErr: false,
		},
		{
			name: "default",
			args: args{
				flags: &androidFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
						AppID:    "com.example.test",
					},
					Keystore:   "./testdata/my.keystore",
					TargetArch: &targetArchFlag{string(ArchMultiple)},
				},
			},
			want: []Context{
				{
					AppBuild:     "1",
					AppID:        "com.example.test",
					ID:           androidOS,
					OS:           androidOS,
					Architecture: ArchMultiple,
					DockerImage:  androidImage,
					Volume:       vol,
					CacheEnabled: true,
					StripDebug:   true,
					Package:      ".",
					Keystore:     "/app/testdata/my.keystore",
				},
			},
			wantErr: false,
		},
		{
			name: "appID is mandatory",
			args: args{
				flags: &androidFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
					},
					Keystore:   "./testdata/my.keystore",
					TargetArch: &targetArchFlag{string(ArchMultiple)},
				},
			},
			want:    []Context{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := makeAndroidContext(tt.args.flags, tt.args.args)

			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.want, ctx)
		})
	}
}
