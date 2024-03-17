package gateway

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway/http"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkgetcd "github.com/faustuzas/occa/src/pkg/etcd"
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
	Etcd       pkgetcd.Configuration           `yaml:"etcd"`
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts gateway server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	services, err := p.StartServices()
	if err != nil {
		return err
	}
	defer func() {
		if err = services.CloseWithTimeout(15 * time.Second); err != nil {
			p.Logger.Error("error while closing services", zap.Error(err))
		}
	}()

	routes, err := http.Configure(http.Services{
		UsersRegisterer:     services.AuthRegisterer,
		AuthMiddleware:      services.HTTPAuthMiddleware,
		ActiveUsersTracker:  services.ActiveUserTracker,
		EventServerSelector: services.EventServerRegistry,
		Logger:              p.Logger,
		Registry:            services.MetricsRegistry,
	})
	if err != nil {
		return fmt.Errorf("configuring routes: %v", err)
	}

	httpListener, err := p.HTTPListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to address: %w", err)
	}

	defer func() {
		_ = httpListener.Close()
	}()

	var (
		srv      = pkghttp.NewServer(p.Logger, httpListener, routes)
		srvErrCh = make(chan error, 1)
	)
	go func() {
		p.Logger.Info("starting server", zap.Stringer("address", p.HTTPListenAddress))

		srvErrCh <- srv.Start()
	}()

	select {
	case <-p.CloseCh:
		p.Logger.Info("received close request, terminating")
		if err = srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shuting down server: %w", err)
		}
	case err = <-srvErrCh:
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}
