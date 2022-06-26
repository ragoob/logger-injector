package loggerInjector

import (
	"context"
	"encoding/json"
	"fmt"
	models "github.com/ragoob/logger-injector/models"
	utils "github.com/ragoob/logger-injector/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Injector struct {
	client *utils.Client
	Kind   string
}

func (i *Injector) Inject(ctx context.Context, result *models.Result, config *utils.Config) error {
	if !ensureCanAddSideCar(result) {
		return fmt.Errorf("logger sidecar already injected  to [%s-%s]", result.Namespace, result.Name)
	}
	if len(result.Spec.Volumes) == 0 {
		return fmt.Errorf("the object should contain at least one volume")
	}
	annotation, err := utils.NewConfigInstanceFromAnnotation(result.Name, result.Spec.Volumes[0].Name, result.Annotations)
	if err != nil {
		return err
	}

	return i.injectFluentdContainer(ctx, result, annotation, config)
}

func (i *Injector) injectFluentdContainer(ctx context.Context, result *models.Result, annotation *utils.Annotation, config *utils.Config) error {
	configMap, err := createFluentdConfigMap(ctx, i.client, result.Namespace, result.Name, annotation, config)
	if err != nil {
		return err
	}
	mainContainer := result.Spec.Containers[0]
	result.Spec.Containers = append(result.Spec.Containers, createSideCareContainerObject(mainContainer, result.Name, config))
	fluentdVolume, err := createFluentdVolumeObject(ctx, result.Namespace, result.Name, annotation, i.client)
	if err != nil {
		return err
	}
	result.Spec.Volumes = append(result.Spec.Volumes, fluentdVolume)
	fluentdConfigVolume, err := createFluentdConfigMapVolumeObject(configMap)
	if err != nil {
		return err
	}
	result.Spec.Volumes = append(result.Spec.Volumes, fluentdConfigVolume)
	result.Labels[utils.InjectorInjectedAnnotation] = "true"

	bytes, err := i.preparePatchPayload(result)
	if err != nil {
		return err
	}

	_, patchErr := i.client.Patch(ctx, result.Namespace, result.Name, i.Kind, bytes, v1.PatchOptions{})
	if patchErr != nil {
		return patchErr
	}
	return nil
}
func (i *Injector) preparePatchPayload(result *models.Result) ([]byte, error) {
	if i.Kind == utils.CronJob {
		payload := models.PatchPayload[models.CronJobSpec]{
			Spec: models.CronJobSpec{
				JobTemplate: models.JobTemplate{
					Spec: models.Spec{
						Template: models.Template{
							Spec: result.Spec,
							MetaData: models.MetaData{
								Labels:      result.Labels,
								Annotations: result.Annotations,
							},
						},
					},
				},
			},
		}
		return json.Marshal(payload)
	} else {
		payload := models.PatchPayload[models.Spec]{
			Spec: models.Spec{
				Template: models.Template{
					Spec: result.Spec,
					MetaData: models.MetaData{
						Labels:      result.Labels,
						Annotations: result.Annotations,
					},
				},
			},
		}
		return json.Marshal(payload)
	}
}
