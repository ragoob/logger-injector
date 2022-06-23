package loggerInjector

import (
	"context"
	"errors"
	utils "github.com/ragoob/logger-injector/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	meta1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch_type "k8s.io/apimachinery/pkg/watch"
	"sync"
	"time"
)

type Watcher struct {
	sync.Mutex
	client utils.Client
}

func NewWatcher() *Watcher {
	client := *utils.NewClient()
	return &Watcher{
		client: client,
	}
}

func (w *Watcher) watchLoop(ctx context.Context) error {
	injector := &Injector{
		client: w.client,
	}
	watch, err := w.client.Instance.AppsV1().Deployments(CoreV1.NamespaceAll).Watch(ctx, meta1.ListOptions{})
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
					log.Fatal("unexpected type")
					error <- errors.New("unexpected type")
				}

				obj, ok := event.Object.(*v1.Deployment)
				if !ok {
					log.Fatal("unexpected type")
					error <- errors.New("unexpected type")
				}
				annotations := obj.Spec.Template.GetObjectMeta().GetAnnotations()

				if (event.Type == watch_type.Added || event.Type == watch_type.Modified) && annotations[utils.InjectorAgentAnnotation] != "" {
					w.Lock()
					err := injector.Inject(ctx, obj)
					if err != nil {
						log.Errorf("failed to inject sidecar for [%s] pod with error : [%s]", obj.Name, err.Error())
					}
					w.Unlock()

				}

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

func (w *Watcher) Watch(ctx context.Context) error {
	for {
		if err := w.watchLoop(ctx); err != nil {
			logrus.Error(err)
		}
	}
}
