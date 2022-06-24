package main

import (
	"context"
	loggerInjector "github.com/ragoob/logger-injector/services"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("SideCar logger injector worker started!")
	watcher := loggerInjector.NewWatcher()
	//start watch k8s
	watcher.Watch(context.Background())
}
