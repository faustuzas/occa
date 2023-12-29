package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"

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

	logger, err := config.Logger.GetService()
	if err != nil {
		fmt.Printf("failed to configure logger: %v\n", err)
		return
	}

	pkglaunch.WaitForInterrupt(logger, func(closeCh <-chan struct{}) error {
		return gateway.Start(gateway.Params{
			Configuration: config,
			Logger:        logger,
			Registry:      prometheus.NewRegistry(),
			CloseCh:       closeCh,
		})
	})
}
