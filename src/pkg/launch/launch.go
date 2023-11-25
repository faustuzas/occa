package launch

import (
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
)

type ServiceLaunch func(closeCh <-chan struct{}) error

func WaitForInterrupt(logger *zap.Logger, launch ServiceLaunch) {
	var (
		closeCh = make(chan struct{})
		errCh   = make(chan error, 1)
	)

	go func() {
		errCh <- launch(closeCh)
	}()

	select {
	case err := <-errCh:
		logger.Error("terminated with error", zap.Error(err))
		return
	case <-interruptCh():
		close(closeCh)
	}

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("service shutdown with error", zap.Error(err))
		}
	case <-time.After(15 * time.Second):
		logger.Error("application failed to terminate gracefully in 15s")
	}
}

func interruptCh() <-chan struct{} {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	resultCh := make(chan struct{})
	go func() {
		<-ch
		close(resultCh)
	}()

	return resultCh
}
