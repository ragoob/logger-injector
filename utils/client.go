package loggerInjector

import (
	"context"
	"flag"
	"fmt"
	CoreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
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

func (c *Client) Patch(ctx context.Context, nameSpace string, name string, kind string, data []byte, options metaV1.PatchOptions) (runtime.Object, error) {
	switch kind {
	case Deployment:
		return c.Instance.AppsV1().Deployments(nameSpace).Patch(ctx, name, types.MergePatchType, data, options)
	case Stateful:
		return c.Instance.AppsV1().StatefulSets(nameSpace).Patch(ctx, name, types.MergePatchType, data, options)
	default:
		return nil, fmt.Errorf("the %s type is not supported", kind)
	}
}

func (c *Client) GetWatcher(ctx context.Context, kind string, opts metaV1.ListOptions) (watch.Interface, error) {
	switch kind {
	case Deployment:
		return c.Instance.AppsV1().Deployments(CoreV1.NamespaceAll).Watch(ctx, opts)
	case Stateful:
		return c.Instance.AppsV1().StatefulSets(CoreV1.NamespaceAll).Watch(ctx, opts)
	default:
		return nil, fmt.Errorf("the %s type is not supported", kind)
	}
}
