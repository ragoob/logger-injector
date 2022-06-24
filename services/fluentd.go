package loggerInjector

import (
	"context"
	"fmt"
	loggerInjector "github.com/ragoob/logger-injector/utils"
	utils "github.com/ragoob/logger-injector/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateFluentdConfigMap(ctx context.Context, deployment *v1.Deployment, client *loggerInjector.Client, config *loggerInjector.Config) (*CoreV1.ConfigMap, error) {
	name := fmt.Sprintf("%s-fluentd", deployment.Name)
	if existing, err := client.Instance.CoreV1().ConfigMaps(deployment.Namespace).Get(ctx, name,
		meta1.GetOptions{}); err == nil && existing != nil {
		log.Infof("Deleting exsisting configMap [%s]", name)
		err := client.Instance.CoreV1().ConfigMaps(deployment.Namespace).Delete(ctx, name, meta1.DeleteOptions{})
		if err != nil {
			log.Errorf("failed to delete configmap [%s]", name)
			return nil, err
		}
	}

	data, err := createDataObject(config)
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
			Namespace: deployment.Namespace,
		},
		Data: data,
	}
	result, err := client.Instance.CoreV1().ConfigMaps(deployment.Namespace).Create(ctx, &configMap, meta1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createDataObject(config *loggerInjector.Config) (map[string]string, error) {
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
           `, "/var/log/td-agent/"+config.InjectorLogPathPattern,
		"/var/log/td-agent/"+config.InjectorLogPathPattern+".pos",
		config.InjectorLogTag,
		config.Elastic.Host,
		config.Elastic.Port,
		config.Elastic.User,
		config.Elastic.Password,
		config.Elastic.SslVersion,
		config.Elastic.Scheme,
		config.Elastic.SslVerify,
		config.InjectorFlushInterval,
	)

	data[utils.FluentDConfigData] = conf
	return data, nil
}

func createSideCareContainerObject(deployment *v1.Deployment, config *loggerInjector.Config) CoreV1.Container {
	container := CoreV1.Container{}
	container.Name = fmt.Sprintf("%s-fluentd-logger", deployment.Name)
	container.ImagePullPolicy = deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy
	container.Image = config.FluentdImageRepository
	container.VolumeMounts = append(container.VolumeMounts, CoreV1.VolumeMount{
		Name:      deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name,
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

func createFluentdVolumeObject(ctx context.Context, deployment *v1.Deployment, config *loggerInjector.Config, client *loggerInjector.Client) (CoreV1.Volume, error) {
	volume := CoreV1.Volume{}
	volume.PersistentVolumeClaim = &CoreV1.PersistentVolumeClaimVolumeSource{}
	volume.Name = utils.FluentdBufferVolumeName
	if config.InjectorStorageClassName != "" {
		pvc, err := createFluentdPvc(ctx, deployment, config, client)
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

func createFluentdPvc(ctx context.Context, deployment *v1.Deployment, config *loggerInjector.Config, client *loggerInjector.Client) (*CoreV1.PersistentVolumeClaim, error) {
	pvc := &CoreV1.PersistentVolumeClaim{
		Spec: CoreV1.PersistentVolumeClaimSpec{},
	}
	name := fmt.Sprintf("fluentd-log-%s", deployment.Name)
	existing, getErr := client.Instance.CoreV1().PersistentVolumeClaims(deployment.Namespace).Get(ctx, name, meta1.GetOptions{})
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
	pvc, err := client.Instance.CoreV1().PersistentVolumeClaims(deployment.Namespace).Create(ctx, pvc, meta1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return pvc, nil
}
