//go:build k8s
// +build k8s

package command

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/fyne-io/fyne-cross/internal/cloud"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"
)

type kubernetesContainerEngine struct {
	baseEngine

	aws    *cloud.AWSSession
	client *cloud.K8sClient

	mutex        sync.Mutex
	currentImage *kubernetesContainerImage

	namespace    string
	s3Path       string
	storageLimit string

	noProjectUpload  bool
	noResultDownload bool
}

var client *cloud.K8sClient

func kubernetesFlagSet(flagSet *flag.FlagSet, flags *CommonFlags) {
	flagSet.BoolVar(&flags.NoProjectUpload, "no-project-upload", false, "Will reuse the project data available in S3, used by the kubernetes engine.")
	flagSet.BoolVar(&flags.NoResultDownload, "no-result-download", false, "Will not download the result of the compilation from S3 automatically, used by the kubernetes engine.")
	flagSet.StringVar(&flags.Namespace, "namespace", "default", "The namespace the kubernetes engine will use to run the pods in, used by and imply the kubernetes engine.")
	flagSet.StringVar(&flags.S3Path, "S3-path", "/", "The path to push to and pull data from, used by the kubernetes engine.")
	flagSet.StringVar(&flags.SizeLimit, "size-limit", "2Gi", "The size limit of mounted filesystem inside the container, used by the kubernetes engine.")
}

func checkKubernetesClient() (err error) {
	client, err = cloud.GetKubernetesClient()
	return err
}

func newKubernetesContainerRunner(context Context) (containerEngine, error) {
	aws, err := cloud.NewAWSSessionFromEnvironment()
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, err
	}

	engine := &kubernetesContainerEngine{
		baseEngine: baseEngine{
			env:  context.Env,
			tags: context.Tags,
			vol:  context.Volume,
		},
		namespace:        context.Namespace,
		s3Path:           context.S3Path,
		noProjectUpload:  context.NoProjectUpload,
		noResultDownload: context.NoResultDownload,
		aws:              aws,
		client:           client,
		storageLimit:     context.SizeLimit,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Interupting")
		engine.close()
		os.Exit(1)
	}()

	return engine, nil
}

type kubernetesContainerImage struct {
	baseContainerImage

	pod *cloud.Pod

	noProjectUpload bool

	cloudLocalMount []containerMountPoint

	runner *kubernetesContainerEngine
}

var _ containerEngine = (*kubernetesContainerEngine)(nil)
var _ closer = (*kubernetesContainerImage)(nil)

func (r *kubernetesContainerEngine) createContainerImage(arch Architecture, OS string, image string) containerImage {
	noProjectUpload := r.noProjectUpload
	r.noProjectUpload = false // We only need to upload data to S3 from the host once.

	return r.createContainerImageInternal(arch, OS, image, func(base baseContainerImage) containerImage {
		ret := &kubernetesContainerImage{
			baseContainerImage: base,
			noProjectUpload:    noProjectUpload,
			runner:             r,
		}
		ret.cloudLocalMount = append(ret.cloudLocalMount, containerMountPoint{name: "cache", inContainer: r.vol.CacheDirContainer()})
		return ret
	})
}

func (r *kubernetesContainerEngine) close() {
	r.mutex.Lock()
	currentImage := r.currentImage
	r.mutex.Unlock()

	if currentImage != nil {
		currentImage.close()
	}
}

func (i *kubernetesContainerImage) Engine() containerEngine {
	return i.runner
}

func (i *kubernetesContainerImage) close() error {
	var err error

	defer i.runner.mutex.Unlock()
	i.runner.mutex.Lock()

	if i.pod != nil {
		err = i.pod.Close()
		i.pod = nil
	}
	if i.runner.currentImage == i {
		i.runner.currentImage = nil
	}

	return err
}

func (i *kubernetesContainerImage) Run(vol volume.Volume, opts options, cmdArgs []string) error {
	return i.pod.Run(context.Background(), opts.WorkDir, cmdArgs)
}

func AddAWSParameters(aws *cloud.AWSSession, command string, s ...string) []string {
	r := []string{command}

	if endpoint := os.Getenv("AWS_S3_ENDPOINT"); endpoint != "" {
		r = append(r, "--aws-endpoint", endpoint)
	}
	if region := os.Getenv("AWS_S3_REGION"); region != "" {
		r = append(r, "--aws-region", region)
	}
	if bucket := os.Getenv("AWS_S3_BUCKET"); bucket != "" {
		r = append(r, "--aws-bucket", bucket)
	}

	creds, err := aws.GetCredentials()
	if err == nil {
		if creds.AccessKeyID != "" {
			r = append(r, "--aws-AKID", creds.AccessKeyID)
		}
		if creds.SecretAccessKey != "" {
			r = append(r, "--aws-secret", creds.SecretAccessKey)
		}
	} else {
		log.Infof("Impossible to get AWS credentials.")
	}

	return append(r, s...)
}

func appendKubernetesEnv(env []cloud.Env, environs map[string]string) []cloud.Env {
	for k, v := range environs {
		env = append(env, cloud.Env{Name: k, Value: v})
	}
	return env
}

func (i *kubernetesContainerImage) Prepare() error {
	var err error

	// Upload all mount point to S3
	if !i.noProjectUpload {
		log.Infof("Uploading project to S3")
		for _, mountPoint := range i.mount {
			log.Infof("Uploading directory %s compressed to %s.", mountPoint.localHost, i.runner.s3Path+"/"+mountPoint.name+".tar.zstd")
			err = i.runner.aws.UploadCompressedDirectory(mountPoint.localHost, i.runner.s3Path+"/"+mountPoint.name+".tar.zstd")
			if err != nil {
				return err
			}
		}
	}

	// Build pod
	var mount []cloud.Mount

	for _, mountPoint := range append(i.mount, i.cloudLocalMount...) {
		mount = append(mount, cloud.Mount{
			Name:            mountPoint.name,
			PathInContainer: mountPoint.inContainer,
		})
	}

	cgo := "1"
	if i.os == webOS {
		cgo = "0"
	}
	env := []cloud.Env{
		{Name: "CGO_ENABLED", Value: cgo},                            // enable CGO
		{Name: "GOCACHE", Value: i.runner.vol.GoCacheDirContainer()}, // mount GOCACHE to reuse cache between builds
	}
	env = appendKubernetesEnv(env, i.runner.env)
	env = appendKubernetesEnv(env, i.env)

	// This allow to run more than one fyne-cross for a specific architecture per cluster namespace
	var unique [6]byte
	rand.Read(unique[:])

	name := fmt.Sprintf("fyne-cross-%s-%x", i.ID(), unique)
	namespace := i.runner.namespace

	i.pod, err = i.runner.client.NewPod(context.Background(),
		name, i.DockerImage, namespace,
		i.runner.storageLimit, i.runner.vol.WorkDirContainer(), mount, env)
	if err != nil {
		return err
	}

	download := func(vol volume.Volume, downloadPath string, containerPath string) error {
		log.Infof("Downloading %s to %s", downloadPath, containerPath)
		return i.Run(i.runner.vol, options{},
			AddAWSParameters(i.runner.aws, "fyne-cross-s3", "download-directory", downloadPath, containerPath),
		)
	}

	// Download data from S3 for all mount point
	log.Infof("Downloading project content from S3 into Kubernetes cluster")
	for _, mountPoint := range i.mount {
		err = download(i.runner.vol, i.runner.s3Path+"/"+mountPoint.name+".tar.zstd", mountPoint.inContainer)
		if err != nil {
			return err
		}
	}
	log.Infof("Download cached data if available from S3 into Kubernetes cluster")
	for _, mountPoint := range i.cloudLocalMount {
		download(i.runner.vol, i.runner.s3Path+"/"+mountPoint.name+"-"+i.ID()+".tar.zstd", mountPoint.inContainer)
	}

	log.Infof("Done preparing pods")

	i.runner.currentImage = i

	return nil
}

func (i *kubernetesContainerImage) Finalize(packageName string) (ret error) {
	// Terminate pod on exit
	defer func() {
		err := i.close()
		if err != nil {
			ret = err
		}
	}()

	// golang does use a LIFO for defer, this will be triggered before the i.close()
	defer i.runner.mutex.Unlock()
	i.runner.mutex.Lock()

	// Upload package result to S3
	uploadPath := i.runner.s3Path + "/" + i.ID() + "/" + packageName
	log.Infof("Uploading package %s to S3", packageName)
	// Darwin application are actually directory and we need
	// to compress it in a format that Darwin understand by default
	if strings.ToLower(filepath.Ext(packageName)) == ".app" {
		uploadPath += ".tar.xz"
		ret = i.Run(i.runner.vol, options{},
			AddAWSParameters(i.runner.aws,
				"fyne-cross-s3", "upload-directory",
				volume.JoinPathContainer(i.runner.vol.TmpDirContainer(), i.ID(), packageName), uploadPath),
		)
	} else {
		ret = i.Run(i.runner.vol, options{},
			AddAWSParameters(i.runner.aws,
				"fyne-cross-s3", "upload-file",
				volume.JoinPathContainer(i.runner.vol.TmpDirContainer(), i.ID(), packageName), uploadPath),
		)
	}
	if ret != nil {
		return
	}

	// Upload cached data to S3
	for _, mountPoint := range i.cloudLocalMount {
		log.Infof("Uploading %s to %s", mountPoint.inContainer, i.runner.s3Path+"/"+mountPoint.name+"-"+i.ID()+".tar.zstd")
		err := i.Run(i.runner.vol, options{},
			AddAWSParameters(i.runner.aws,
				"fyne-cross-s3", "upload-directory",
				mountPoint.inContainer, i.runner.s3Path+"/"+mountPoint.name+"-"+i.ID()+".tar.zstd"),
		)
		if err != nil {
			log.Infof("Failed to upload %s", mountPoint.inContainer)
		}
	}

	if !i.runner.noResultDownload {
		// Download package result from S3 locally
		distFile := volume.JoinPathHost(i.runner.vol.DistDirHost(), i.ID(), packageName)
		err := os.MkdirAll(filepath.Dir(distFile), 0755)
		if err != nil {
			ret = fmt.Errorf("could not create the dist package dir: %v", err)
			return
		}

		log.Infof("Downloading result package to %s.", distFile)
		ret = i.runner.aws.DownloadFile(uploadPath, distFile)

		log.Infof("[✓] Package: %q", distFile)
	} else {
		log.Infof("[✓] Package available at : %q", uploadPath)
	}
	return
}
