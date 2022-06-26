package main

import (
	"context"
	loggerInjector "github.com/ragoob/logger-injector/services"
	utils "github.com/ragoob/logger-injector/utils"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("SideCar logger injector worker started!")
	config, err := utils.NewConfig()
	if err != nil {
		log.Fatalf("error load configurations [%v]", err)
	}
	ctx := context.Background()
	loggerInjector.WatchAll(ctx, config)

	<-ctx.Done()
	 
}
