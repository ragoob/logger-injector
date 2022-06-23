package main

import (
	"context"
	"github.com/joho/godotenv"
	loggerInjector "github.com/ragoob/logger-injector/services"
	log "github.com/sirupsen/logrus"
)

func main() {
	//load .env if it exists
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Info("SideCar logger injector worker started! \n")
	watcher := loggerInjector.NewWatcher()
	//start watch k8s
	watcher.Watch(context.Background())
}
