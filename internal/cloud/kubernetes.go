package cloud

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetKubernetesClient() (*rest.Config, *kubernetes.Clientset, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	var config *rest.Config
	var err error

	if !Exists(kubeconfig) {
		// No configuration file, try probing in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Try to build cluster configuration from file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, err
		}
	}

	kubectl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return config, kubectl, nil
}
