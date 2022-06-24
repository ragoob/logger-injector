package loggerInjector

import (
	"flag"
	log "github.com/sirupsen/logrus"
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

func NewClient() *Client {
	var config *rest.Config
	if ConvertToBooleanOrDefault(os.Getenv(InClusterConfig)) {
		_config, err := rest.InClusterConfig()
		if err != nil {
			if err != nil {
				log.Errorf("failed to create k8s config from in cluster configuration [%s]", err.Error())
			}
		}

		config = _config

	} else {
		_config, err := fromKubeConfig()
		if err != nil {
			log.Errorf("failed to create client from kube config [%] \n", err.Error())
			log.Warning("daemon will exit \n")
			os.Exit(2)
		}

		config = _config
	}
	clientSet, _ := kubernetes.NewForConfig(config)
	return &Client{
		Instance: clientSet,
	}

}

func fromKubeConfig() (*restClient.Config, error) {
	var kubeConfig *string
	if os.Getenv(KubeConfigPathEnv) != "" {
		kubeConfig = flag.String("kubeconfig", os.Getenv(KubeConfigPathEnv), "(optional) absolute path to the kubeconfig file")
	} else {
		if home, err := os.UserHomeDir(); home != "" && err == nil {
			kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			if kubeConfig != nil {
				if err != nil {
					log.Errorf("failed to find k8s config in dir [%s] [%s] \n", filepath.Join(home, ".kube", "config"), err.Error())
					panic(err)
				}
			}
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}
