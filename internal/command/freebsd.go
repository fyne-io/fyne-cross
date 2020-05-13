package command

import (
	"fmt"
	"runtime"

	"github.com/lucor/fyne-cross/v2/internal/log"
)

const (
	// freebsdOS it the freebsd OS name
	freebsdOS = "freebsd"
	// freebsdImageAmd64 is the fyne-cross image for the FreeBSD OS amd64 arch
	freebsdImageAmd64 = "lucor/fyne-cross:freebsd-latest"
)

var (
	// freebsdArchSupported defines the supported target architectures on freebsd
	freebsdArchSupported = []Architecture{ArchAmd64}
)

// FreeBSD build and package the fyne app for the freebsd OS
type FreeBSD struct {
	Context []Context
}

// Name returns the one word command name
func (cmd *FreeBSD) Name() string {
	return "freebsd"
}

// Description returns the command description
func (cmd *FreeBSD) Description() string {
	return "Build and package a fyne application for the freebsd OS"
}

// Parse parses the arguments and set the usage for the command
func (cmd *FreeBSD) Parse(args []string) error {
	commonFlags, err := newCommonFlags()
	if err != nil {
		return err
	}

	flags := &freebsdFlags{
		CommonFlags: commonFlags,
		TargetArch:  &targetArchFlag{runtime.GOARCH},
	}
	flagSet.Var(flags.TargetArch, "arch", fmt.Sprintf(`List of target architecture to build separated by comma. Supported arch: %s`, freebsdArchSupported))

	flagSet.Usage = cmd.Usage
	flagSet.Parse(args)

	ctx, err := freebsdContext(flags, flagSet.Args())
	if err != nil {
		return err
	}
	cmd.Context = ctx
	return nil
}

// Run runs the command
func (cmd *FreeBSD) Run() error {

	for _, ctx := range cmd.Context {

		log.Infof("[i] Target: %s/%s", ctx.OS, ctx.Architecture)
		log.Debugf("%#v", ctx)

		//
		// prepare build
		//
		err := cleanTargetDirs(ctx)
		if err != nil {
			return err
		}

		err = goModInit(ctx)
		if err != nil {
			return err
		}

		//
		// build
		//
		err = goBuild(ctx)
		if err != nil {
			return err
		}

		//
		// package
		//
		log.Infof("[!] Packaging for %s not implemented yet", ctx.OS)
	}

	return nil
}

// Usage displays the command usage
func (cmd *FreeBSD) Usage() {
	data := struct {
		Name        string
		Description string
	}{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}

	template := `
Usage: fyne-cross {{ .Name }} [options] [package]

{{ .Description }}

Options:
`

	printUsage(template, data)
	flagSet.PrintDefaults()
}

// freebsdFlags defines the command-line flags for the freebsd command
type freebsdFlags struct {
	*CommonFlags

	// TargetArch represents a list of target architecture to build on separated by comma
	TargetArch *targetArchFlag
}

// freebsdContext returns the command context for a freebsd target
func freebsdContext(flags *freebsdFlags, args []string) ([]Context, error) {

	targetArch, err := targetArchFromFlag(*flags.TargetArch, freebsdArchSupported)
	if err != nil {
		return []Context{}, fmt.Errorf("could not make build context for %s OS: %s", freebsdOS, err)
	}

	ctxs := []Context{}
	for _, arch := range targetArch {

		ctx, err := makeDefaultContext(flags.CommonFlags, args)
		if err != nil {
			return ctxs, err
		}

		ctx.Architecture = arch
		ctx.OS = freebsdOS
		ctx.ID = fmt.Sprintf("%s-%s", ctx.OS, ctx.Architecture)

		var defaultDockerImage string
		switch arch {
		case ArchAmd64:
			defaultDockerImage = freebsdImageAmd64
			ctx.Env = append(ctx.Env, "GOOS=freebsd", "GOARCH=amd64", "CC=x86_64-unknown-freebsd11-clang")
		}

		// set context based on command-line flags
		if flags.DockerImage == "" {
			ctx.DockerImage = defaultDockerImage
		}

		ctxs = append(ctxs, ctx)
	}

	return ctxs, nil
}
