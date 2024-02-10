package main

import (
	"flag"
	"fmt"

	"github.com/faustuzas/occa/src/eventserver"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkglaunch "github.com/faustuzas/occa/src/pkg/launch"
)

var configFile = flag.String("f", "deploy/config/event-server.yml", "configuration file")

func main() {
	flag.Parse()

	config, err := pkgconfig.LoadConfig[eventserver.Configuration](*configFile)
	if err != nil {
		fmt.Printf("failed to load configuration: %v\n", err)
		return
	}

	logger, err := config.Logger.Build()
	if err != nil {
		fmt.Printf("failed to configure logger: %v\n", err)
		return
	}

	pkglaunch.WaitForInterrupt(logger, func(closeCh <-chan struct{}) error {
		return eventserver.Start(eventserver.Params{
			Configuration: config,
			Logger:        logger,
			CloseCh:       closeCh,
		})
	})
}
