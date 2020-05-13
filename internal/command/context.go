package command

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lucor/fyne-cross/v2/internal/log"
	"github.com/lucor/fyne-cross/v2/internal/volume"
)

const (
	// ArchAmd64 represents the amd64 architecture
	ArchAmd64 Architecture = "amd64"
	// Arch386 represents the amd64 architecture
	Arch386 Architecture = "386"
	// ArchArm64 represents the arm64 architecture
	ArchArm64 Architecture = "arm64"
	// ArchArm represents the arm architecture
	ArchArm Architecture = "arm"
)

// Architecture represents the Architecture type
type Architecture string

func (a Architecture) String() string {
	return (string)(a)
}

// Context represent a build context
type Context struct {
	// Volume holds the mounted volumes between host and containers
	volume.Volume

	Architecture          // Arch defines the target architecture
	Env          []string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	ID           string   // ID is the context ID
	LdFlags      []string // LdFlags defines the ldflags to use
	OS           string   // OS defines the target OS
	Tags         []string // Tags defines the tags to use

	AppID        string // AppID is the appID to use for distribution
	CacheEnabled bool   // CacheEnabled if true enable go build cache
	DockerImage  string // DockerImage defines the docker image used to build
	Icon         string // Icon is the optional icon in png format to use for distribution
	Output       string // Output is the name output
	Package      string // Package is the package to build named by the import path as per 'go build'
	StripDebug   bool   // StripDebug if true, strips binary output
	Debug        bool   // Debug if true enable debug log
}

// String implements the Stringer interface
func (ctx Context) String() string {
	buf := &bytes.Buffer{}

	template := `
Architecture: {{ .Architecture }}
OS: {{ .OS }}
Output: {{ .Output }}
`

	log.PrintTemplate(buf, template, ctx)
	return buf.String()
}

func makeDefaultContext(flags *CommonFlags) (Context, error) {
	// mount the fyne-cross volume
	vol, err := volume.Mount(flags.RootDir, flags.CacheDir)
	if err != nil {
		return Context{}, err
	}

	// set context based on command-line flags
	ctx := Context{
		AppID:        flags.AppID,
		CacheEnabled: !flags.NoCache,
		DockerImage:  flags.DockerImage,
		Env:          flags.Env,
		Icon:         flags.Icon,
		Output:       flags.Output,
		StripDebug:   !flags.NoStripDebug,
		Debug:        flags.Debug,
		Volume:       vol,
	}

	ctx.Package, err = packageFromFlag(flags.Package, vol)
	if err != nil {
		return ctx, err
	}

	if len(flags.Ldflags) > 0 {
		ctx.LdFlags = append(ctx.LdFlags, flags.Ldflags)
	}

	if flags.Silent {
		log.SetLevel(log.LevelSilent)
	}

	if flags.Debug {
		log.SetLevel(log.LevelDebug)
	}

	return ctx, nil
}

// packageFromFlag validates and returns the package to compile.
func packageFromFlag(pkg string, vol volume.Volume) (string, error) {
	if pkg == "." || pkg == "" {
		return ".", nil
	}

	if !filepath.IsAbs(pkg) {
		return pkg, nil
	}

	rel, err := filepath.Rel(vol.WorkDirContainer(), pkg)
	if err != nil || strings.HasPrefix(rel, "..") {
		fmt.Println(rel)
		return pkg, fmt.Errorf("package options when specified as absolute path must be relative to the project root dir")
	}

	return pkg, nil
}

// targetArchFromFlag validates and returns the architecture specified using flag againts the supported ones.
// If flagVar contains the wildcard char "*" all the supported architecture are returned.
func targetArchFromFlag(flagVar []string, supportedArch []Architecture) ([]Architecture, error) {
	targetArch := []Architecture{}
Loop:
	for _, v := range flagVar {
		if v == "*" {
			return supportedArch, nil
		}
		for _, valid := range supportedArch {
			if Architecture(v) == valid {
				targetArch = append(targetArch, valid)
				continue Loop
			}
		}
		return nil, fmt.Errorf("arch %q is not supported. Supported: %s", v, supportedArch)
	}
	return targetArch, nil
}
