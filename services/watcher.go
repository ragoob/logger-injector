package loggerInjector

import (
	"context"
	"fmt"
	models "github.com/ragoob/logger-injector/models"
	utils "github.com/ragoob/logger-injector/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchType "k8s.io/apimachinery/pkg/watch"
	"os"
	"sync"
	"time"
)

type Watcher struct {
	sync.Mutex
	client   *utils.Client
	injector *Injector
	config   *utils.Config
}

func newWatcher(kind string, config *utils.Config, client *utils.Client) *Watcher {
	return &Watcher{
		client: client,
		config: config,
		injector: &Injector{
			client: client,
			Kind:   kind,
		},
	}

}

func (w *Watcher) watchLoop(ctx context.Context) error {
	log.Infof("Start watching loop for kind [%s]", w.injector.Kind)
	watch, err := w.client.GetWatcher(ctx, w.injector.Kind, metaV1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	channel := watch.ResultChan()
	done := make(chan bool, 1)
	error := make(chan error, 1)
	defer close(done)
	go func() {
		for {
			select {
			case event, ok := <-channel:
				if !ok {
					log.Errorf("unexpected type error in watching [%s]", w.injector.Kind)
					error <- fmt.Errorf("unexpected type")
				}

				obj, ok := w.ensureObject(event)

				if !ok {
					log.Errorf("unexpected type error in watching [%s]", w.injector.Kind)
					error <- fmt.Errorf("unexpected type")
				}

				if IsReadForInjection(event, obj) {
					err := w.injector.Inject(ctx, obj, w.config)
					if err != nil {
						log.Errorf("failed to inject sidecar for [%s] pod with error : [%s]", obj.Name, err.Error())
					} else {
						log.Infof("Sidecar logger sucessfully injected to [%s-%s]", obj.Namespace, obj.Name)
					}
				}
				//K8s watcher default session is 30 min, so we restart the go-routine every 30 min
			case <-time.After(30 * time.Minute):
				logrus.Info("Timeout, restarting event watcher")
				done <- true

			}

		}
	}()

	select {
	case <-done:
		return nil
	case err := <-error:
		return err
	}

}
func IsReadForInjection(event watchType.Event, obj *models.Result) bool {
	return utils.ConvertToBooleanOrDefault(obj.Annotations[utils.InjectorAgentAnnotation]) &&
		(event.Type == watchType.Added || event.Type == watchType.Modified)
}
func (w *Watcher) watch(rootCtx context.Context) {
	for {
		if err := w.watchLoop(rootCtx); err != nil {
			logrus.Errorf("error in watching loop [%v]", err)
		}
	}
}

func (w *Watcher) ensureObject(event watchType.Event) (*models.Result, bool) {
	switch w.injector.Kind {
	case utils.Deployment:
		obj, ok := event.Object.(*v1.Deployment)
		if !ok {
			return nil, ok
		}
		var conditions []models.Condition
		for _, v := range obj.Status.Conditions {
			conditions = append(conditions, models.Condition{
				Status: v.Status,
				Type:   string(v.Type),
			})
		}
		return &models.Result{
			Name:        obj.Name,
			Namespace:   obj.Namespace,
			Spec:        &obj.Spec.Template.Spec,
			Annotations: obj.Spec.Template.GetObjectMeta().GetAnnotations(),
			Labels:      obj.Spec.Template.GetObjectMeta().GetLabels(),
			Conditions:  conditions,
		}, ok
	case utils.Stateful:
		obj, ok := event.Object.(*v1.StatefulSet)
		if !ok {
			return nil, ok
		}
		var conditions []models.Condition
		for _, v := range obj.Status.Conditions {
			conditions = append(conditions, models.Condition{
				Status: v.Status,
				Type:   string(v.Type),
			})
		}
		return &models.Result{
			Name:        obj.Name,
			Namespace:   obj.Namespace,
			Spec:        &obj.Spec.Template.Spec,
			Annotations: obj.Spec.Template.GetObjectMeta().GetAnnotations(),
			Labels:      obj.Spec.Template.GetObjectMeta().GetLabels(),
			Conditions:  conditions,
		}, ok

	default:
		return nil, false
	}
}

func WatchAll(ctx context.Context, config *utils.Config) {
	client, err := utils.NewClient()
	if err != nil {
		log.Fatalf("failed to create K8s client instance [%v]", err)
	}
	if err != nil {
		log.Errorf("failed to create K8s client [%v]", err)
		log.Warning("daemon will exit")
		os.Exit(1)
	}

	resources := []string{utils.Deployment, utils.Stateful}
	for _, v := range resources {
		watcher := newWatcher(v, config, client)
		go watcher.watch(ctx)
	}
}
