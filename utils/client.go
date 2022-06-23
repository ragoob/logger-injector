package loggerInjector

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type Client struct {
	Instance *kubernetes.Clientset
}

func NewClient() *Client {
	var kubeConfig *string
	if home, err := os.UserHomeDir(); home != "" && err == nil {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		if kubeConfig != nil {

			if err != nil {
				log.Errorf("failed to find k8s config in dir [%s] [%s] \n", filepath.Join(home, ".kube", "config"), err.Error())
				panic(err)
			}
		}
	}
	var config *restclient.Config
	if kubeConfig != nil {
		_config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			log.Errorf("failed to create k8s config from cmd [%s]", err.Error())
		}
		config = _config
	} else {
		_config, err := rest.InClusterConfig()
		if err != nil {
			if err != nil {
				log.Errorf("failed to create k8s config from in cluster configuration [%s]", err.Error())
			}
		}
		config = _config
	}

	clientSet, _ := kubernetes.NewForConfig(config)
	return &Client{
		Instance: clientSet,
	}
}
