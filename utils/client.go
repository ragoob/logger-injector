package loggerInjector

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type Client struct {
	Instance *kubernetes.Clientset
}

func NewClient() (*Client, error) {
	var config *rest.Config
	if ConvertToBooleanOrDefault(os.Getenv(InClusterConfig)) {
		_config, err := rest.InClusterConfig()
		if err != nil {
			if err != nil {
				return nil, err
			}
		}

		config = _config

	} else {
		_config, err := fromKubeConfig()
		if err != nil {
			return nil, err
		}

		config = _config
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		Instance: clientSet,
	}, nil

}

func fromKubeConfig() (*restClient.Config, error) {
	var kubeConfig *string
	if home, err := os.UserHomeDir(); home != "" && err == nil {
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		if kubeConfig != nil {
			if err != nil {
				return nil, err
			}
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}
