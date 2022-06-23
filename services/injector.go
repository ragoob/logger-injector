package loggerInjector

import (
	"context"
	"fmt"
	loggerInjector "github.com/ragoob/logger-injector/utils"
	utils "github.com/ragoob/logger-injector/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	meta1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Injector struct {
	client utils.Client
}

func (i *Injector) Inject(ctx context.Context, deployment *v1.Deployment) error {
	if utils.ConvertToBooleanOrDefault(deployment.Spec.Template.GetObjectMeta().GetLabels()[utils.InjectorInjectedAnnotation]) || ifSideCarExists(deployment) {
		log.Debugf("logger sidecar already injected  to [%s-%s]", deployment.Namespace, deployment.Name)
		return nil
	}
	config, err := utils.NewConfigInstanceFromAnnotation(deployment)
	if err != nil {
		return err
	}
	configMap, err := CreateFluentdConfigMap(ctx, deployment, i.client, config)
	if err != nil {
		log.Errorf("error creating configMap: [%s]", err.Error())
	}
	return i.injectFluentdContainer(ctx, deployment, config, configMap)
}

func (i *Injector) injectFluentdContainer(ctx context.Context, deployment *v1.Deployment, config *loggerInjector.Config, configMap *CoreV1.ConfigMap) error {

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, createSideCareContainerObject(deployment, config))
	fluentdVolume, err := createFluentdVolumeObject(ctx, deployment, config, i.client)
	if err != nil {
		return err
	}
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, fluentdVolume)
	fluentdConfigVolume, err := createFluentdConfigMapVolumeObject(configMap)
	if err != nil {
		return err
	}
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, fluentdConfigVolume)
	deployment.Spec.Template.Labels[utils.InjectorInjectedAnnotation] = "true"
	_, updateErr := i.client.Instance.AppsV1().Deployments(deployment.Namespace).Update(context.TODO(), deployment, meta1.UpdateOptions{})
	if updateErr == nil {
		log.Infof("logger sideCar injected successfuly to [%s-%s]", deployment.Namespace, deployment.Name)
	}
	return updateErr
}

func ifSideCarExists(deployment *v1.Deployment) bool {
	name := fmt.Sprintf("%s-fluentd-logger", deployment.Name)

	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == name {

			return true
		}
	}
	return false
}
