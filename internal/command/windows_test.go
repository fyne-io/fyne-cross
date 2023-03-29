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
		wantImages  []containerImage
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
			},
			wantImages: []containerImage{
				&localContainerImage{
					baseContainerImage: baseContainerImage{
						arch: "amd64",
						os:   "windows",
						id:   "windows-amd64",
						env:  map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows", "CXX": "zig c++ -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows"},
						mount: []containerMountPoint{
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
			wantImages: []containerImage{
				&localContainerImage{
					baseContainerImage: baseContainerImage{
						arch: "386",
						os:   "windows",
						id:   "windows-386",
						env:  map[string]string{"GOOS": "windows", "GOARCH": "386", "CC": "zig cc -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows", "CXX": "zig c++ -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows"},
						mount: []containerMountPoint{
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
				Env: map[string]string{
					"GOFLAGS": "-ldflags=-X -ldflags=main.version=1.2.3",
				},
			},
			wantImages: []containerImage{
				&localContainerImage{
					baseContainerImage: baseContainerImage{
						arch: "amd64",
						os:   "windows",
						id:   "windows-amd64",
						env:  map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows", "CXX": "zig c++ -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows"},
						mount: []containerMountPoint{
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
			},
			wantImages: []containerImage{
				&localContainerImage{
					baseContainerImage: baseContainerImage{
						arch: "amd64",
						os:   "windows",
						id:   "windows-amd64",
						env:  map[string]string{"GOOS": "windows", "GOARCH": "amd64", "CC": "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows", "CXX": "zig c++ -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows"},
						mount: []containerMountPoint{
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

				err := windows.setupContainerImages(tt.args.flags, tt.args.args)
				if tt.wantErr {
					require.NotNil(t, err)
					return
				}
				require.Nil(t, err)
				assert.Equal(t, tt.wantContext, windows.defaultContext)

				for index := range windows.Images {
					windows.Images[index].(*localContainerImage).runner = nil
				}

				assert.Equal(t, tt.wantImages, windows.Images)
			})
		})
	}
}
