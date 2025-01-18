package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"
)

type Mount struct {
	Name            string
	PathInContainer string
}

type Env struct {
	Name  string
	Value string
}

type K8sClient struct {
	config  *rest.Config
	kubectl *kubernetes.Clientset
}

type Pod struct {
	client    *K8sClient
	pod       *core.Pod
	name      string
	namespace string
	workDir   string
}

var Log func(string, ...interface{})

func GetKubernetesClient() (*K8sClient, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	var config *rest.Config
	var err error

	if !Exists(kubeconfig) {
		// No configuration file, try probing in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// Try to build cluster configuration from file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	kubectl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sClient{config: config, kubectl: kubectl}, nil
}

func (k8s *K8sClient) NewPod(ctx context.Context, name string, image string, namespace string, storageLimit string, workDir string, mount []Mount, environs []Env) (*Pod, error) {
	var volumesMount []core.VolumeMount
	var volumes []core.Volume

	quantity := resource.MustParse(storageLimit)

	for _, mountPoint := range mount {
		volumesMount = append(volumesMount, core.VolumeMount{
			Name:      mountPoint.Name,
			MountPath: mountPoint.PathInContainer,
		})
		volumes = append(volumes, core.Volume{
			Name: mountPoint.Name,
			VolumeSource: core.VolumeSource{EmptyDir: &core.EmptyDirVolumeSource{
				SizeLimit: &quantity,
			}},
		})
	}

	var envVar []core.EnvVar

	for _, env := range environs {
		envVar = append(envVar, core.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{Name: name,
			Labels: map[string]string{
				"app": "fyne-cross",
			}},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Containers: []core.Container{
				{
					Name:            "fyne-cross",
					Image:           image,
					ImagePullPolicy: core.PullAlways,
					Command:         []string{"/bin/bash"},
					// The pod will stop itself after 30min
					Args:         []string{"-c", "trap : TERM INT; sleep 1800 & wait"},
					Env:          envVar,
					VolumeMounts: volumesMount,
					WorkingDir:   workDir,
				},
			},
			Volumes: volumes,
			Affinity: &core.Affinity{
				PodAntiAffinity: &core.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
						{
							LabelSelector: &meta.LabelSelector{
								MatchExpressions: []meta.LabelSelectorRequirement{
									{
										Key:      "app",
										Operator: "In",
										Values:   []string{"fyne-cross"},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		},
	}

	logWrapper("Creating pod %q: %v", name, pod)

	// Start pod
	api := k8s.kubectl.CoreV1()

	instance, err := api.Pods(namespace).Create(
		ctx, pod, meta.CreateOptions{},
	)
	if err != nil {
		return nil, err
	}

	logWrapper("Waiting for pod to be ready")
	err = wait.PollUntilContextTimeout(ctx, time.Second, time.Duration(10)*time.Minute, true, func(ctx context.Context) (bool, error) {
		pod, err := api.Pods(namespace).Get(ctx, name, meta.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case core.PodRunning:
			return true, nil
		case core.PodFailed, core.PodSucceeded:
			return false, fmt.Errorf("pod terminated")
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &Pod{
		client:    k8s,
		pod:       instance,
		name:      name,
		namespace: namespace,
		workDir:   workDir,
	}, nil
}

func (p *Pod) Run(ctx context.Context, workDir string, cmdArgs []string) error {
	api := p.client.kubectl.CoreV1()

	if workDir != "" && workDir != p.workDir {
		shellCommand := "cd " + workDir + " && " + cmdArgs[0] + " "

		for _, s := range cmdArgs[1:] {
			shellCommand = shellCommand + fmt.Sprintf("%q ", s)
		}

		cmdArgs = []string{
			"sh",
			"-c",
			shellCommand,
		}
	}

	req := api.RESTClient().Post().Resource("pods").Name(p.name).
		Namespace(p.namespace).SubResource("exec")
	option := &core.PodExecOptions{
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
	exec, err := remotecommand.NewSPDYExecutor(p.client.config, "POST", req.URL())
	if err != nil {
		return err
	}

	logWrapper("Executing command %v", cmdArgs)
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stderr,
		Stderr: os.Stderr,
		Tty:    false,
	})
	return err
}

func (p *Pod) Close() error {
	deletePolicy := meta.DeletePropagationForeground
	if err := p.client.kubectl.CoreV1().Pods(p.namespace).Delete(context.Background(), p.name, meta.DeleteOptions{
		PropagationPolicy: &deletePolicy}); err != nil {
		return err
	}
	return nil
}

func logWrapper(format string, p ...interface{}) {
	if Log == nil {
		return
	}
	Log(format, p...)
}
