package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkglaunch "github.com/faustuzas/occa/src/pkg/launch"
)

var configFile = flag.String("f", "", "configuration file")

func main() {
	flag.Parse()

	if len(*configFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	config, err := pkgconfig.LoadConfig[gateway.Configuration](*configFile)
	if err != nil {
		fmt.Printf("failed to load configuration: %v\n", err)
		return
	}

	logger, err := config.Logger.Build()
	if err != nil {
		fmt.Printf("failed to configure logger: %v\n", err)
		return
	}
	logger = logger.With(zap.String("component", "gateway"))

	pkglaunch.WaitForInterrupt(logger, func(closeCh <-chan struct{}) error {
		return gateway.Start(gateway.Params{
			Configuration: config,
			Logger:        logger,
			CloseCh:       closeCh,
		})
	})
}
