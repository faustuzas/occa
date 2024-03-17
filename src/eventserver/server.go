package eventserver

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	esgrpc "github.com/faustuzas/occa/src/eventserver/grpc"
	"github.com/faustuzas/occa/src/eventserver/http"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkgetcd "github.com/faustuzas/occa/src/pkg/etcd"
	esmembership "github.com/faustuzas/occa/src/pkg/eventserver/membership"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	Auth pkgauth.ValidatorConfiguration `yaml:"auth"`

	HTTPListenAddress *pkgnet.ListenAddr `yaml:"httpListenAddress"`
	GRPCListenAddress *pkgnet.ListenAddr `yaml:"grpcListenAddress"`

	Etcd     pkgetcd.Configuration     `yaml:"etcd"`
	MemStore pkgmemstore.Configuration `yaml:"memstore"`

	ServerID string `yaml:"serverID"`
}

func (c Configuration) Validate() error {
	if c.ServerID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}
	return nil
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts chat server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	if err := p.Configuration.Validate(); err != nil {
		return fmt.Errorf("validating configuration: %w", err)
	}

	services, err := p.StartServices()
	if err != nil {
		return err
	}

	serverStarted := false
	defer func() {
		if !serverStarted {
			if e := services.CloseWithTimeout(15 * time.Second); e != nil {
				p.Logger.Error("error while closing services", zap.Error(e))
			}
		}
	}()

	httpListener, err := p.HTTPListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to http address: %w", err)
	}

	httpHandler, err := http.Configure(http.Services{
		EventServer:        services.EventServer,
		HTTPAuthMiddleware: services.HTTPAuthMiddleware,
		Logger:             p.Logger,
		Registry:           services.MetricsRegistry,
	})
	if err != nil {
		return fmt.Errorf("configuring http handler: %w", err)
	}
	httpServer := pkghttp.NewServer(p.Logger, httpListener, httpHandler)

	grpcListener, err := p.GRPCListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to gRPC address: %w", err)
	}

	var grpcServer *grpc.Server
	if grpcServer, err = esgrpc.Configure(esgrpc.Services{
		EventServer:           services.EventServer,
		StreamAuthInterceptor: services.GRPCStreamAuthInterceptor,
		Instrumentation: pkginstrument.Instrumentation{
			Logger:     p.Logger,
			Registerer: services.MetricsRegistry,
		},
	}); err != nil {
		return fmt.Errorf("configuring GRPC server: %w", err)
	}

	// Create a close channel where various actors might trigger termination.
	// Make it relatively big so no one would get blocked.
	closeCh := make(chan error, 10)
	go func() {
		p.Logger.Info("starting HTTP server", zap.Stringer("address", p.HTTPListenAddress))
		closeCh <- httpServer.Start()
	}()

	go func() {
		p.Logger.Info("starting gRPC server", zap.Stringer("address", p.GRPCListenAddress))
		closeCh <- grpcServer.Serve(grpcListener)
	}()

	lostLeaseC, err := services.MembershipManager.JoinCluster(context.Background(),
		func(ctx context.Context) (esmembership.ServerInfo, error) {
			return esmembership.ServerInfo{
				ID:          p.ServerID,
				GRPCAddress: p.Configuration.GRPCListenAddress.String(),
				HTTPAddress: p.Configuration.HTTPListenAddress.String(),
			}, nil
		})
	if err != nil {
		return fmt.Errorf("joining cluster: %w", err)
	}
	go func() {
		select {
		case <-lostLeaseC:
			closeCh <- fmt.Errorf("lost membership in the cluster")
		case <-closeCh:
		}
	}()

	go func() {
		select {
		case <-p.CloseCh:
			p.Logger.Info("received close request, terminating")
			closeCh <- nil
		case <-closeCh:
		}
	}()

	serverStarted = true

	// wait until first termination trigger
	closeErr := <-closeCh

	closeCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err = services.EventServer.InitiateShutdown(closeCtx); err != nil {
		p.Logger.Error("failed initiation event server shutdown", zap.Error(err))
	}

	if err = services.MembershipManager.LeaveCluster(closeCtx); err != nil {
		p.Logger.Error("failed leaving cluster", zap.Error(err))
	}

	g, _ := errgroup.WithContext(closeCtx)
	g.Go(func() error {
		grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error {
		if err = httpServer.Shutdown(closeCtx); err != nil {
			p.Logger.Error("failed closing HTTP server", zap.Error(err))
		}
		return nil
	})

	_ = g.Wait()

	if err = services.Close(closeCtx); err != nil {
		p.Logger.Error("error while closing services", zap.Error(err))
	}

	return closeErr
}
