package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
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

	logger, err := config.BuildLogger()
	if err != nil {
		fmt.Printf("failed to configure logger: %v\n", err)
		return
	}
	logger = logger.With(zap.String("component", "gateway"))

	var (
		closeCh = make(chan struct{})
		errCh   = make(chan error, 1)
	)

	go func() {
		errCh <- gateway.Start(gateway.Params{
			Configuration: config,
			Logger:        logger,
			CloseCh:       closeCh,
		})
	}()

	select {
	case err = <-errCh:
		logger.Error("terminated with error", zap.Error(err))
		return
	case <-WaitForInterrupt():
		close(closeCh)
	}

	select {
	case err = <-errCh:
		if err != nil {
			logger.Error("server shutdown with error", zap.Error(err))
		}
	case <-time.After(15 * time.Second):
		logger.Error("application failed to terminate gracefully in 15s")
	}
}

func WaitForInterrupt() <-chan struct{} {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	resultCh := make(chan struct{})
	go func() {
		<-ch
		close(resultCh)
	}()

	return resultCh
}
