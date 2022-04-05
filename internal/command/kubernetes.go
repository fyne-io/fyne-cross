package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

type KubernetesContainerRunner struct {
	AllContainerRunner

	aws     *cloud.AWSSession
	config  *rest.Config
	kubectl *kubernetes.Clientset

	namespace    string
	s3Path       string
	storageLimit resource.Quantity

	noProjectUpload bool
}

func NewKubernetesContainerRunner(context Context) (ContainerRunner, error) {
	aws, err := cloud.NewAWSSessionFromEnvironment()
	if err != nil {
		return nil, err
	}

	config, kubectl, err := cloud.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	return &KubernetesContainerRunner{
		AllContainerRunner: AllContainerRunner{
			Env:   context.Env,
			Tags:  context.Tags,
			vol:   context.Volume,
			debug: context.Debug,
		},
		namespace:       context.Namespace,
		s3Path:          context.S3Path,
		noProjectUpload: context.NoProjectUpload,
		aws:             aws,
		kubectl:         kubectl,
		config:          config,
		storageLimit:    resource.MustParse(context.SizeLimit),
	}, nil
}

type KubernetesContainerImage struct {
	AllContainerImage

	podName string

	noProjectUpload bool

	CloudLocalMount []ContainerMountPoint

	Runner *KubernetesContainerRunner
	pod    *v1.Pod
}

var _ io.Closer = (*KubernetesContainerImage)(nil)

func (r *KubernetesContainerRunner) NewImageContainer(arch Architecture, OS string, image string) ContainerImage {
	noProjectUpload := r.noProjectUpload
	r.noProjectUpload = false // We only need to upload data to S3 from the host once.

	return r.newImageContainerInternal(arch, OS, image, func(arch Architecture, OS, ID, image string) ContainerImage {
		ret := &KubernetesContainerImage{
			AllContainerImage: AllContainerImage{
				Architecture: arch,
				OS:           OS,
				ID:           ID,
				DockerImage:  image,
				Env:          make(map[string]string),
			},
			noProjectUpload: noProjectUpload,
			Runner:          r,
		}
		ret.CloudLocalMount = append(ret.CloudLocalMount, ContainerMountPoint{Name: "cache", InContainer: r.vol.CacheDirContainer()})
		return ret
	})
}

func (i *KubernetesContainerImage) GetRunner() ContainerRunner {
	return i.Runner
}

func (i *KubernetesContainerImage) Close() error {
	if i.podName != "" {
		deletePolicy := metav1.DeletePropagationForeground
		if err := i.Runner.kubectl.CoreV1().Pods(i.Runner.namespace).Delete(context.Background(), i.podName, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy}); err != nil {
			return err
		}
	}

	return nil
}

func (i *KubernetesContainerImage) Run(vol volume.Volume, opts Options, cmdArgs []string) error {
	api := i.Runner.kubectl.CoreV1()

	if opts.WorkDir != "" && opts.WorkDir != i.Runner.vol.WorkDirContainer() {
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
		Namespace(i.Runner.namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmdArgs,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(i.Runner.config, "POST", req.URL())
	if err != nil {
		return err
	}

	log.Infof("Executing command %v", cmdArgs)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
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

func (i *KubernetesContainerImage) Prepare() error {
	var err error

	// Upload all mount point to S3
	if !i.noProjectUpload {
		log.Infof("Uploading project to S3")
		for _, mountPoint := range i.Mount {
			log.Infof("Uploading directory %s compressed to %s.", mountPoint.LocalHost, i.Runner.s3Path+"/"+mountPoint.Name+".tar.zstd")
			err = i.Runner.aws.UploadCompressedDirectory(mountPoint.LocalHost, i.Runner.s3Path+"/"+mountPoint.Name+".tar.zstd")
			if err != nil {
				return err
			}
		}
	}

	// Build pod
	var volumesMount []v1.VolumeMount
	var volumes []v1.Volume

	for _, mountPoint := range append(i.Mount, i.CloudLocalMount...) {
		volumesMount = append(volumesMount, v1.VolumeMount{
			Name:      mountPoint.Name,
			MountPath: mountPoint.InContainer,
		})
		volumes = append(volumes, v1.Volume{
			Name: mountPoint.Name,
			VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{
				SizeLimit: &i.Runner.storageLimit,
			}},
		})
	}

	i.podName = fmt.Sprintf("fyne-cross-%s", i.GetID())
	namespace := i.Runner.namespace
	timeout := time.Duration(10) * time.Minute

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: i.podName},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "fyne-crosss",
					Image:           i.DockerImage,
					ImagePullPolicy: v1.PullAlways,
					Command:         []string{"/bin/bash"},
					// The pod will stop itself after 30min
					Args:         []string{"-c", "trap : TERM INT; sleep 1800 & wait"},
					VolumeMounts: volumesMount,
					WorkingDir:   i.Runner.vol.WorkDirContainer(),
				},
			},
			Volumes: volumes,
		},
	}

	log.Infof("Creating pod %q", i.podName)

	// Start pod
	api := i.Runner.kubectl.CoreV1()

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
		return i.Run(i.Runner.vol, Options{},
			AddAWSParameters(i.Runner.aws, "fyne-cross-s3", "download-directory", downloadPath, containerPath),
		)
	}

	// Download data from S3 for all mount point
	log.Infof("Downloading project content from S3 into Kubernetes cluster")
	for _, mountPoint := range i.Mount {
		err = download(i.Runner.vol, i.Runner.s3Path+"/"+mountPoint.Name+".tar.zstd", mountPoint.InContainer)
		if err != nil {
			return err
		}
	}
	log.Infof("Download cached data if available from S3 into Kubernetes cluster")
	for _, mountPoint := range i.CloudLocalMount {
		download(i.Runner.vol, i.Runner.s3Path+"/"+mountPoint.Name+"-"+i.ID+".tar.zstd", mountPoint.InContainer)
	}

	return nil
}

func (i *KubernetesContainerImage) Finalize(packageName string) (ret error) {
	// Terminate pod on exit
	defer func() {
		err := i.Close()
		if err != nil {
			ret = err
		}
	}()

	// Upload package result to S3
	log.Infof("Uploading package %s to S3", packageName)
	ret = i.Run(i.Runner.vol, Options{},
		AddAWSParameters(i.Runner.aws,
			"fyne-cross-s3", "upload-file",
			volume.JoinPathContainer(i.Runner.vol.TmpDirContainer(), i.GetID(), packageName), i.Runner.s3Path+"/"+packageName),
	)
	if ret != nil {
		return
	}

	// Upload cached data to S3
	for _, mountPoint := range i.CloudLocalMount {
		log.Infof("Uploading %s to %s", mountPoint.InContainer, i.Runner.s3Path+"/"+mountPoint.Name+"-"+i.ID+".tar.zstd")
		err := i.Run(i.Runner.vol, Options{},
			AddAWSParameters(i.Runner.aws,
				"fyne-cross-s3", "upload-directory",
				mountPoint.InContainer, i.Runner.s3Path+"/"+mountPoint.Name+"-"+i.ID+".tar.zstd"),
		)
		if err != nil {
			log.Infof("Failed to upload %s", mountPoint.InContainer)
		}
	}

	// Download package result from S3 locally
	distFile := volume.JoinPathHost(i.Runner.vol.DistDirHost(), i.GetID(), packageName)
	err := os.MkdirAll(filepath.Dir(distFile), 0755)
	if err != nil {
		ret = fmt.Errorf("could not create the dist package dir: %v", err)
		return
	}

	log.Infof("Downloading result package to %s.", distFile)
	ret = i.Runner.aws.DownloadFile(i.Runner.s3Path+"/"+packageName, distFile)

	log.Infof("[✓] Package: %s", distFile)
	return
}
