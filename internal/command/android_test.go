package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_makeAndroidContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	engine, err := MakeEngine(autodetectEngine)
	if err != nil {
		t.Skip("engine not found", err)
	}

	type args struct {
		flags *androidFlags
		args  []string
	}
	tests := []struct {
		name        string
		args        args
		wantContext Context
		wantImages  []ContainerImage
		wantErr     bool
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
			wantContext: Context{},
			wantImages:  []ContainerImage{},
			wantErr:     true,
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
			wantContext: Context{},
			wantImages:  []ContainerImage{},
			wantErr:     true,
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
			wantContext: Context{
				AppBuild:     "1",
				AppID:        "com.example.test",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Keystore:     "/app/testdata/my.keystore",
				Engine:       engine,
				Env:          map[string]string{},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: ArchMultiple,
						OS:           androidOS,
						ID:           androidOS,
						Env:          map[string]string{},
						Mount: []ContainerMountPoint{
							{vol.WorkDirHost(), vol.WorkDirContainer()},
							{vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: androidImage,
					},
					Pull: false,
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
			wantContext: Context{
				AppBuild:     "1",
				AppID:        "com.example.test",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Keystore:     "/app/testdata/my.keystore",
				Engine:       engine,
				Env:          map[string]string{},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: ArchMultiple,
						OS:           androidOS,
						ID:           androidOS,
						Env:          map[string]string{},
						Mount: []ContainerMountPoint{
							{vol.WorkDirHost(), vol.WorkDirContainer()},
							{vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: androidImage,
					},
					Pull: false,
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
			wantContext: Context{},
			wantImages:  []ContainerImage{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			android := NewAndroidCommand()

			err := android.makeAndroidContainerImages(tt.args.flags, tt.args.args)

			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tt.wantContext, android.defaultContext)

			for index := range android.Images {
				android.Images[index].(*LocalContainerImage).Runner = nil
			}

			assert.Equal(t, tt.wantImages, android.Images)
		})
	}
}
