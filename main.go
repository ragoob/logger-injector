package main

import (
	"context"
	"github.com/joho/godotenv"
	loggerInjector "github.com/ragoob/logger-injector/services"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	watcher := loggerInjector.NewWatcher()

	watcher.Watch(context.Background())
}
