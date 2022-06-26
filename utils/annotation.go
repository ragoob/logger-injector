package loggerInjector

import (
	"fmt"
)

type Annotation struct {
	InjectorClaimName        string `json:"logger.injector.io/app-claim-name"`
	InjectorLogTag           string `json:"logger.injector.io/log-tag-name"`
	InjectorFlushInterval    string `json:"logger.injector.io/flush-interval"`
	InjectorLogPathPattern   string `json:"logger.injector.io/log-path-pattern"`
	InjectorStorageClassName string `json:"logger.injector.io/storage-class-name"`
	FluentdVolumeSize        string `json:"logger.injector.io/fluentd-vol-size"`
}

func NewConfigInstanceFromAnnotation(objectName string, volumeName string, annotations map[string]string) (*Annotation, error) {
	var config Annotation
	configMap, err := createConfigOrDefault(objectName, volumeName, annotations)
	if err != nil {
		return nil, err
	}
	parseErr := MapToStruct(configMap, &config)
	if parseErr != nil {
		return nil, parseErr
	}
	return &config, nil
}
func createConfigOrDefault(objectName string, volumeName string, annotations map[string]string) (map[string]interface{}, error) {
	config, err := getDefaultConfig(objectName, volumeName)
	if err != nil {
		return nil, err
	}
	for k := range config {
		if annotations[k] != "" {
			config[k] = annotations[k]
		}
	}
	return config, nil
}
func getDefaultConfig(objectName string, volumeName string) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	configMap[InjectorAgentAnnotation] = true
	configMap[InjectorClaimName] = volumeName
	configMap[InjectorLogTag] = fmt.Sprintf("log.%s", objectName)
	configMap[InjectorFlushInterval] = InjectorFlushIntervalDefault
	configMap[InjectorLogPathPattern] = InjectorLogPathPatternDefault
	configMap[InjectorStorageClassName] = ""
	configMap[FluentdVolumeSize] = FluentdVolumeSizeDefault
	return configMap, nil
}
