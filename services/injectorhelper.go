package loggerInjector

import (
	"context"
	"fmt"

	models "github.com/ragoob/logger-injector/models"
	loggerInjector "github.com/ragoob/logger-injector/utils"
	utils "github.com/ragoob/logger-injector/utils"
	log "github.com/sirupsen/logrus"
	CoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateFluentdConfigMap(ctx context.Context, nameSpace string, objectName string, client *loggerInjector.Client, annotation *loggerInjector.Annotation, config *utils.Config) (*CoreV1.ConfigMap, error) {
	name := fmt.Sprintf("%s-fluentd", objectName)
	if existing, err := client.Instance.CoreV1().ConfigMaps(nameSpace).Get(ctx, name,
		meta1.GetOptions{}); err == nil && existing != nil {
		return nil, fmt.Errorf("[%s] resource already exist", name)
	}

	data, err := createDataObject(annotation, config)
	if err != nil {
		return nil, err
	}
	configMap := CoreV1.ConfigMap{
		TypeMeta: meta1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: meta1.ObjectMeta{
			Name:      name,
			Namespace: nameSpace,
		},
		Data: data,
	}
	result, err := client.Instance.CoreV1().ConfigMaps(nameSpace).Create(ctx, &configMap, meta1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createDataObject(annotation *loggerInjector.Annotation, config *utils.Config) (map[string]string, error) {
	data := make(map[string]string)
	conf := fmt.Sprintf(`
               <source>
        @type tail
        path  %s
        pos_file %s
        tag  %s
        <parse>
         @type json
        </parse>
      </source>

      <match **>
        @id elasticsearch
        @type elasticsearch
        include_tag_key true
        type_name log
        host  %s
        port  %v
        user  %s
        password %s
        ssl_version %s
        scheme %s
        ssl_verify %v
        index_name ${tag}
        logstash_format true
        logstash_prefix ${tag}
        flush_interval %s
      </match> 
           `, "/var/log/td-agent/"+annotation.InjectorLogPathPattern,
		"/var/log/td-agent/"+annotation.InjectorLogPathPattern+".pos",
		annotation.InjectorLogTag,
		config.Elastic.Host,
		config.Elastic.Port,
		config.Elastic.User,
		config.Elastic.Password,
		config.Elastic.SslVersion,
		config.Elastic.Scheme,
		config.Elastic.SslVerify,
		annotation.InjectorFlushInterval,
	)

	data[utils.FluentDConfigData] = conf
	return data, nil
}

func createSideCareContainerObject(mainContainer CoreV1.Container, objectName string, config *loggerInjector.Config) CoreV1.Container {
	container := CoreV1.Container{}
	container.Name = fmt.Sprintf("%s-fluentd-logger", objectName)
	container.ImagePullPolicy = mainContainer.ImagePullPolicy
	container.Image = config.FluentdImageRepository
	container.Env = append(container.Env, CoreV1.EnvVar{
		Name:  "FLUENTD_ARGS",
		Value: "-c /fluentd/etc/fluent.conf",
	})
	container.VolumeMounts = append(container.VolumeMounts, CoreV1.VolumeMount{
		Name:      mainContainer.VolumeMounts[0].Name,
		MountPath: utils.FluentdLogPath,
	})
	container.VolumeMounts = append(container.VolumeMounts, CoreV1.VolumeMount{
		Name:      utils.FluentdBufferVolumeName,
		MountPath: utils.FluentdBufferPath,
	})
	container.VolumeMounts = append(container.VolumeMounts, CoreV1.VolumeMount{
		Name:      utils.FluentdConfigMapVolumeName,
		MountPath: utils.FluentdConfigPath,
	})
	return container
}

func createFluentdVolumeObject(ctx context.Context, nameSpace string, objectName string, config *loggerInjector.Annotation, client *loggerInjector.Client) (CoreV1.Volume, error) {
	volume := CoreV1.Volume{}
	volume.PersistentVolumeClaim = &CoreV1.PersistentVolumeClaimVolumeSource{}
	volume.Name = utils.FluentdBufferVolumeName
	if config.InjectorStorageClassName != "" {
		pvc, err := createFluentdPvc(ctx, nameSpace, objectName, config, client)
		if err != nil {
			return volume, err
		}

		volume.PersistentVolumeClaim.ClaimName = pvc.Name

	} else {

		volume.EmptyDir = &CoreV1.EmptyDirVolumeSource{}
	}
	return volume, nil
}

func createFluentdConfigMapVolumeObject(configMap *CoreV1.ConfigMap) (CoreV1.Volume, error) {
	volume := CoreV1.Volume{}
	mode := int32(utils.FluentdConfigMapVolumeDefaultMode)
	volume.Name = utils.FluentdConfigMapVolumeName
	volume.ConfigMap = &CoreV1.ConfigMapVolumeSource{}
	volume.ConfigMap.Name = configMap.Name
	volume.ConfigMap.DefaultMode = &mode
	return volume, nil
}

func createFluentdPvc(ctx context.Context, nameSpace string, objectName string, config *loggerInjector.Annotation, client *loggerInjector.Client) (*CoreV1.PersistentVolumeClaim, error) {
	pvc := &CoreV1.PersistentVolumeClaim{
		Spec: CoreV1.PersistentVolumeClaimSpec{},
	}
	name := fmt.Sprintf("fluentd-log-%s", objectName)
	existing, getErr := client.Instance.CoreV1().PersistentVolumeClaims(nameSpace).Get(ctx, name, meta1.GetOptions{})
	if getErr == nil && existing != nil {
		return existing, nil
	}
	pvc.Name = name
	pvc.Spec.StorageClassName = &config.InjectorStorageClassName
	pvc.Spec.AccessModes = []CoreV1.PersistentVolumeAccessMode{CoreV1.ReadWriteMany}
	pvc.Spec.Resources = CoreV1.ResourceRequirements{
		Requests: CoreV1.ResourceList{
			CoreV1.ResourceStorage: resource.MustParse(config.FluentdVolumeSize),
		},
	}
	pvc, err := client.Instance.CoreV1().PersistentVolumeClaims(nameSpace).Create(ctx, pvc, meta1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return pvc, nil
}
func ensureCanAddSideCar(result *models.Result) bool {
	ok := true
	name := fmt.Sprintf("%s-fluentd-logger", result.Name)
	isMarkedAsInjected := utils.ConvertToBooleanOrDefault(result.Labels[utils.InjectorInjectedAnnotation])
	if isMarkedAsInjected {
		return false
	}

	for _, c := range result.Spec.Containers {
		if c.Name == name {
			return false
		}
	}

	for _, v := range result.Spec.Volumes {
		if v.Name == utils.FluentdBufferVolumeName || v.Name == utils.FluentdConfigMapVolumeName {
			return false
		}
	}
	return ok
}
func createFluentdConfigMap(ctx context.Context, client *utils.Client, nameSpace string, objectName string, annotation *loggerInjector.Annotation, config *loggerInjector.Config) (*CoreV1.ConfigMap, error) {
	configMap, err := CreateFluentdConfigMap(ctx, nameSpace, objectName, client, annotation, config)
	if err != nil {
		log.Errorf("error creating configMap: [%s]", err.Error())
		return nil, err
	}
	return configMap, nil
}
