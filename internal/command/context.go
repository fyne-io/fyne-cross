package command

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
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
	// ArchMultiple represents the universal architecture used by some OS to
	// identify a binary that supports multiple architectures (fat binary)
	ArchMultiple Architecture = "multiple"
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

	Engine    Engine            // Engine is the container engine to use
	Namespace string            // Namespace used by Kubernetes engine to run its pod in
	S3Path    string            // Project base directory to use to push and pull data from S3
	SizeLimit string            // Container mount point size limits honored by Kubernetes only
	Env       map[string]string // Env is the list of custom env variable to set. Specified as "KEY=VALUE"
	Tags      []string          // Tags defines the tags to use
	Metadata  map[string]string // Metadata contain custom metadata passed to fyne package

	AppBuild         string // Build number
	AppID            string // AppID is the appID to use for distribution
	AppVersion       string // AppVersion is the version number in the form x, x.y or x.y.z semantic version
	CacheEnabled     bool   // CacheEnabled if true enable go build cache
	Icon             string // Icon is the optional icon in png format to use for distribution
	Name             string // Name is the application name
	Package          string // Package is the package to build named by the import path as per 'go build'
	Release          bool   // Enable release mode. If true, prepares an application for public distribution
	StripDebug       bool   // StripDebug if true, strips binary output
	Debug            bool   // Debug if true enable debug log
	Pull             bool   // Pull if true attempts to pull a newer version of the docker image
	NoProjectUpload  bool   // NoProjectUpload if true, the build will be done with the artifact already stored on S3
	NoResultDownload bool   // NoResultDownload if true, the result of the build will be left on S3 and not downloaded locally
	NoNetwork        bool   // NoNetwork if true, the build will be done without network access
	ExtraMount       string //ExtraMount if is set, mount this directories Ex: FROM|TO,FROM2|TO2
	//Build context
	BuildMode string // The -buildmode argument to pass to go build

	// Release context
	Category     string //Category represents the category of the app for store listing [macOS]
	Certificate  string //Certificate represents the name of the certificate to sign the build [iOS, Windows]
	Developer    string //Developer represents the developer identity for your Microsoft store account [Windows]
	Keystore     string //Keystore represents the location of .keystore file containing signing information [Android]
	KeystorePass string //KeystorePass represents the password for the .keystore file [Android]
	KeyPass      string //KeyPass represents the assword for the signer's private key, which is needed if the private key is password-protected [Android]
	KeyName      string //KeyName represents the name of the key to sign the build [Android]
	Password     string //Password represents the password for the certificate used to sign the build [Windows]
	Profile      string //Profile represents the name of the provisioning profile for this release build [iOS]
}

// String implements the Stringer interface
func (ctx Context) String() string {
	buf := &bytes.Buffer{}

	template := `
Architecture: {{ .Architecture }}
OS: {{ .OS }}
Name: {{ .Name }}
`

	log.PrintTemplate(buf, template, ctx)
	return buf.String()
}

func overrideDockerImage(flags *CommonFlags, image string) string {
	if flags.DockerImage != "" {
		return flags.DockerImage
	}

	if flags.DockerRegistry != "" {
		return fmt.Sprintf("%s/%s", flags.DockerRegistry, image)
	}

	return image
}

func makeDefaultContext(flags *CommonFlags, args []string) (Context, error) {
	// mount the fyne-cross volume
	vol, err := volume.Mount(flags.RootDir, flags.CacheDir)
	if err != nil {
		return Context{}, err
	}

	engine := flags.Engine.Engine
	if (engine == Engine{}) {
		if flags.Namespace != "" && flags.Namespace != "default" {
			engine, err = MakeEngine(kubernetesEngine)
			if err != nil {
				return Context{}, err
			}
		} else {
			// attempt to autodetect
			engine, err = MakeEngine(autodetectEngine)
			if err != nil {
				return Context{}, err
			}
		}
	}

	// set context based on command-line flags
	ctx := Context{
		AppID:            flags.AppID,
		AppVersion:       flags.AppVersion,
		CacheEnabled:     !flags.NoCache,
		NoProjectUpload:  flags.NoProjectUpload,
		NoResultDownload: flags.NoResultDownload,
		Engine:           engine,
		Namespace:        flags.Namespace,
		S3Path:           flags.S3Path,
		SizeLimit:        flags.SizeLimit,
		Env:              make(map[string]string),
		Tags:             flags.Tags,
		Metadata:         flags.Metadata.values,
		Icon:             flags.Icon,
		Name:             flags.Name,
		StripDebug:       !flags.NoStripDebug,
		Debug:            flags.Debug,
		NoNetwork:        flags.NoNetwork,
		Volume:           vol,
		Pull:             flags.Pull,
		Release:          flags.Release,
		ExtraMount:       flags.ExtraMount,
	}

	if flags.AppBuild <= 0 {
		return ctx, errors.New("build number should be greater than 0")
	}

	// the flag name that replace the deprecated output should not be used
	// as path. Returns error if contains \ or /
	// Fixes: #9
	// TODO: update the error message once the output flag is removed
	if strings.ContainsAny(flags.Name, "\\/") {
		return ctx, errors.New("output and app name should not be used as path")
	}

	for _, v := range flags.Env {
		parts := strings.SplitN(v, "=", 2)
		ctx.Env[parts[0]] = parts[1]
	}

	ctx.AppBuild = strconv.Itoa(flags.AppBuild)

	ctx.Package, err = packageFromArgs(args, vol)
	if err != nil {
		return ctx, err
	}

	if env := os.Getenv("GOFLAGS"); env != "" {
		ctx.Env["GOFLAGS"] = env
	}

	if len(flags.Ldflags) > 0 {
		goflags := ""
		for _, ldflags := range strings.Fields(flags.Ldflags) {
			goflags += "-ldflags=" + ldflags + " "
		}
		if v, ok := ctx.Env["GOFLAGS"]; ok {
			ctx.Env["GOFLAGS"] = strings.TrimSpace(v + " " + goflags)
		} else {
			ctx.Env["GOFLAGS"] = strings.TrimSpace(goflags)
		}
	}

	if flags.Silent {
		log.SetLevel(log.LevelSilent)
	}

	if flags.Debug {
		log.SetLevel(log.LevelDebug)
		debugEnable = true
	}

	return ctx, nil
}

// packageFromArgs validates and returns the package to compile.
func packageFromArgs(args []string, vol volume.Volume) (string, error) {
	pkg := "."
	if len(args) > 0 {
		pkg = args[0]
	}
	if pkg == "." {
		return ".", nil
	}

	if !filepath.IsAbs(pkg) {
		return pkg, nil
	}

	pkg = filepath.Clean(pkg)

	if !strings.HasPrefix(pkg, vol.WorkDirHost()) {
		return pkg, fmt.Errorf("package options when specified as absolute path must be relative to the project root dir")
	}

	pkg = strings.Replace(pkg, vol.WorkDirHost(), ".", 1)
	if runtime.GOOS == "windows" {
		pkg = filepath.ToSlash(pkg)
	}
	return pkg, nil
}

// targetArchFromFlag validates and returns the architecture specified using flag against the supported ones.
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
