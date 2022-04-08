package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_makeWindowsContext(t *testing.T) {
	vol, err := mockDefaultVolume()
	require.Nil(t, err)

	engine, err := MakeEngine(autodetectEngine)
	if err != nil {
		t.Skip("engine not found", err)
	}

	type args struct {
		flags *windowsFlags
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
			name: "default",
			args: args{
				flags: &windowsFlags{
					CommonFlags: &CommonFlags{
						AppBuild: 1,
					},
					TargetArch: &targetArchFlag{"amd64"},
				},
			},
			wantContext: Context{
				AppBuild:     "1",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Engine:       engine,
				Env:          map[string]string{},
				LdFlags:      []string{"-H=windowsgui"},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: "amd64",
						OS:           "windows",
						ID:           "windows-amd64",
						Env:          map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "x86_64-w64-mingw32-gcc"},
						Mount: []ContainerMountPoint{
							{"project", vol.WorkDirHost(), vol.WorkDirContainer()},
							{"cache", vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: windowsImage,
					},
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
			wantContext: Context{
				AppBuild:     "1",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Engine:       engine,
				Env:          map[string]string{},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: "386",
						OS:           "windows",
						ID:           "windows-386",
						Env:          map[string]string{"GOOS": "windows", "GOARCH": "386", "CC": "i686-w64-mingw32-gcc"},
						Mount: []ContainerMountPoint{
							{"project", vol.WorkDirHost(), vol.WorkDirContainer()},
							{"cache", vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: windowsImage,
					},
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
			wantContext: Context{
				AppBuild:     "1",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Engine:       engine,
				Env:          map[string]string{},
				LdFlags:      []string{"-X main.version=1.2.3", "-H=windowsgui"},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: "amd64",
						OS:           "windows",
						ID:           "windows-amd64",
						Env:          map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "x86_64-w64-mingw32-gcc"},
						Mount: []ContainerMountPoint{
							{"project", vol.WorkDirHost(), vol.WorkDirContainer()},
							{"cache", vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: windowsImage,
					},
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
			wantContext: Context{
				AppBuild:     "1",
				Volume:       vol,
				CacheEnabled: true,
				StripDebug:   true,
				Package:      ".",
				Engine:       engine,
				Env:          map[string]string{},
				LdFlags:      []string{"-H=windowsgui"},
			},
			wantImages: []ContainerImage{
				&LocalContainerImage{
					AllContainerImage: AllContainerImage{
						Architecture: "amd64",
						OS:           "windows",
						ID:           "windows-amd64",
						Env:          map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "x86_64-w64-mingw32-gcc"},
						Mount: []ContainerMountPoint{
							{"project", vol.WorkDirHost(), vol.WorkDirContainer()},
							{"cache", vol.CacheDirHost(), vol.CacheDirContainer()},
						},
						DockerImage: "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				windows := NewWindowsCommand()

				err := windows.makeWindowsContainerImages(tt.args.flags, tt.args.args)
				if tt.wantErr {
					require.NotNil(t, err)
					return
				}
				require.Nil(t, err)
				assert.Equal(t, tt.wantContext, windows.defaultContext)

				for index := range windows.Images {
					windows.Images[index].(*LocalContainerImage).Runner = nil
				}

				assert.Equal(t, tt.wantImages, windows.Images)
			})
		})
	}
}
