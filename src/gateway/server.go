package gateway

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway/http"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	HTTPListenAddress *pkgnet.ListenAddr `yaml:"listenAddress"`

	MemStore   pkgmemstore.Configuration       `yaml:"memstore"`
	Auth       pkgauth.ValidatorConfiguration  `yaml:"auth"`
	Registerer pkgauth.RegistererConfiguration `yaml:"registerer"`
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts gateway server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	var (
		logger   = p.Logger
		registry = prometheus.NewRegistry()
	)

	services, err := p.StartServices()
	if err != nil {
		return err
	}
	defer func() {
		if e := services.Close(); e != nil {
			logger.Error("error while closing services", zap.Error(e))
		}
	}()

	routes, err := http.Configure(http.Services{
		UsersRegisterer:    services.AuthRegisterer,
		AuthMiddleware:     services.HTTPAuthMiddleware,
		ActiveUsersTracker: services.ActiveUserTracker,
		Logger:             p.Logger,
		Registry:           registry,
	})
	if err != nil {
		return fmt.Errorf("configuring routes: %v", err)
	}

	httpListener, err := p.HTTPListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to address: %w", err)
	}

	var (
		srv      = pkghttp.NewServer(logger, httpListener, routes)
		srvErrCh = make(chan error, 1)
	)
	go func() {
		logger.Info("starting server", zap.Stringer("address", p.HTTPListenAddress))

		srvErrCh <- srv.Start()
	}()

	select {
	case <-p.CloseCh:
		logger.Info("received close request, terminating")
		if err = srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shuting down server: %w", err)
		}
	case err = <-srvErrCh:
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}
