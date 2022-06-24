package loggerInjector

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"os"
)

type Config struct {
	Elastic                  Elastic `json:"logger-side-car-elastic"`
	InjectorClaimName        string  `json:"logger.injector.io/app-claim-name"`
	InjectorLogTag           string  `json:"logger.injector.io/log-tag-name"`
	InjectorFlushInterval    string  `json:"logger.injector.io/flush-interval"`
	InjectorLogPathPattern   string  `json:"logger.injector.io/log-path-pattern"`
	InjectorStorageClassName string  `json:"logger.injector.io/storage-class-name"`
	FluentdImageRepository   string  `json:"logger-fluentd-image-repository"`
	FluentdVolumeSize        string  `json:"logger.injector.io/fluentd-vol-size"`
}

func NewConfigInstanceFromAnnotation(deployment *v1.Deployment) (*Config, error) {
	var config Config
	configMap, err := createConfigOrDefault(deployment)
	if err != nil {
		return nil, err
	}
	parseErr := MapToStruct(configMap, &config)
	if parseErr != nil {
		return nil, parseErr
	}
	return &config, nil
}
func createConfigOrDefault(deployment *v1.Deployment) (map[string]interface{}, error) {
	config, err := getDefaultConfig(deployment)
	if err != nil {
		return nil, err
	}
	for k := range config {
		if deployment.Spec.Template.GetObjectMeta().GetAnnotations()[k] != "" {

			config[k] = deployment.Spec.Template.GetObjectMeta().GetAnnotations()[k]
		}
	}
	return config, nil
}

func getDefaultConfig(deployment *v1.Deployment) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	configMap[InjectorAgentAnnotation] = true
	configMap[InjectorElastic] = Elastic{
		Host:       os.Getenv("ELASTIC_HOST"),
		Port:       ConvertToIntOrDefault(os.Getenv("ELASTIC_PORT")),
		Password:   os.Getenv("ELASTIC_PASSWORD"),
		User:       os.Getenv("ELASTIC_USER"),
		SslVerify:  ConvertToBooleanOrDefault(os.Getenv("ELASTIC_SSL_VERIFY")),
		Scheme:     os.Getenv("ELASTIC_SCHEME"),
		SslVersion: os.Getenv("ELASTIC_SSL_VERSION"),
	}
	configMap[FluentdImageRepository] = os.Getenv("FLUENTD_IMAGE_REPOSITORY")
	if len(deployment.Spec.Template.Spec.Volumes) == 0 {
		return nil, errors.New(fmt.Sprintf("the deployment [%s] should  contains at least one volumes", deployment.Name))
	}
	configMap[InjectorClaimName] = deployment.Spec.Template.Spec.Volumes[0].Name
	configMap[InjectorLogTag] = fmt.Sprintf("log.%s", deployment.Name)
	configMap[InjectorFlushInterval] = "1m"
	configMap[InjectorLogPathPattern] = "log*.log"
	configMap[InjectorStorageClassName] = "nfs-server-k8s-1.20"
	configMap[FluentdVolumeSize] = "1Gi"
	return configMap, nil
}
