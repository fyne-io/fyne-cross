package command

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fyne-io/fyne-cross/internal/cloud"
	"github.com/fyne-io/fyne-cross/internal/log"
	"github.com/fyne-io/fyne-cross/internal/volume"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

type kubernetesContainerEngine struct {
	baseEngine

	aws     *cloud.AWSSession
	config  *rest.Config
	kubectl *kubernetes.Clientset

	namespace    string
	s3Path       string
	storageLimit resource.Quantity

	noProjectUpload  bool
	noResultDownload bool
}

func newKubernetesContainerRunner(context Context) (containerEngine, error) {
	aws, err := cloud.NewAWSSessionFromEnvironment()
	if err != nil {
		return nil, err
	}

	config, kubectl, err := cloud.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &kubernetesContainerEngine{
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
		kubectl:          kubectl,
		config:           config,
		storageLimit:     resource.MustParse(context.SizeLimit),
	}, nil
}

type kubernetesContainerImage struct {
	baseContainerImage

	podName string

	noProjectUpload bool

	cloudLocalMount []containerMountPoint

	runner *kubernetesContainerEngine
	pod    *v1.Pod
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

func (i *kubernetesContainerImage) Engine() containerEngine {
	return i.runner
}

func (i *kubernetesContainerImage) close() error {
	if i.podName != "" {
		deletePolicy := metav1.DeletePropagationForeground
		if err := i.runner.kubectl.CoreV1().Pods(i.runner.namespace).Delete(context.Background(), i.podName, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy}); err != nil {
			return err
		}
	}

	return nil
}

func (i *kubernetesContainerImage) Run(vol volume.Volume, opts options, cmdArgs []string) error {
	api := i.runner.kubectl.CoreV1()

	if opts.WorkDir != "" && opts.WorkDir != i.runner.vol.WorkDirContainer() {
		shellCommand := "cd " + opts.WorkDir + " && " + cmdArgs[0] + " "

		for _, s := range cmdArgs[1:] {
			shellCommand = shellCommand + fmt.Sprintf("%q ", s)
		}

		cmdArgs = []string{
			"sh",
			"-c",
			shellCommand,
		}
	}

	req := api.RESTClient().Post().Resource("pods").Name(i.podName).
		Namespace(i.runner.namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmdArgs,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(i.runner.config, "POST", req.URL())
	if err != nil {
		return err
	}

	log.Infof("Executing command %v", cmdArgs)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stderr,
		Stderr: os.Stderr,
		Tty:    false,
	})
	return err
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

func appendKubernetesEnv(env []v1.EnvVar, environs map[string]string) []v1.EnvVar {
	for k, v := range environs {
		env = append(env, v1.EnvVar{Name: k, Value: v})
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
	var volumesMount []v1.VolumeMount
	var volumes []v1.Volume

	for _, mountPoint := range append(i.mount, i.cloudLocalMount...) {
		volumesMount = append(volumesMount, v1.VolumeMount{
			Name:      mountPoint.name,
			MountPath: mountPoint.inContainer,
		})
		volumes = append(volumes, v1.Volume{
			Name: mountPoint.name,
			VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{
				SizeLimit: &i.runner.storageLimit,
			}},
		})
	}

	// This allow to run more than one fyne-cross for a specific architecture per cluster namespace
	var unique [6]byte
	rand.Read(unique[:])

	i.podName = fmt.Sprintf("fyne-cross-%s-%x", i.ID(), unique)
	namespace := i.runner.namespace
	timeout := time.Duration(10) * time.Minute
	env := []v1.EnvVar{
		{Name: "CGO_ENABLED", Value: "1"},                            // enable CGO
		{Name: "GOCACHE", Value: i.runner.vol.GoCacheDirContainer()}, // mount GOCACHE to reuse cache between builds
	}
	env = appendKubernetesEnv(env, i.runner.env)
	env = appendKubernetesEnv(env, i.env)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: i.podName},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "fyne-cross",
					Image:           i.DockerImage,
					ImagePullPolicy: v1.PullAlways,
					Command:         []string{"/bin/bash"},
					// The pod will stop itself after 30min
					Args:         []string{"-c", "trap : TERM INT; sleep 1800 & wait"},
					Env:          env,
					VolumeMounts: volumesMount,
					WorkingDir:   i.runner.vol.WorkDirContainer(),
				},
			},
			Volumes: volumes,
		},
	}

	log.Infof("Creating pod %q", i.podName)
	if debugging() {
		log.Debug(pod)
	}

	// Start pod
	api := i.runner.kubectl.CoreV1()

	i.pod, err = api.Pods(namespace).Create(
		context.Background(),
		pod,
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}

	log.Infof("Waiting for pod to be ready")
	err = wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		pod, err := api.Pods(namespace).Get(context.Background(), i.podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, fmt.Errorf("pod terminated")
		}
		return false, nil
	})
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
