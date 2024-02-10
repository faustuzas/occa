package eventserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"

	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	HTTPListenAddress *pkgnet.ListenAddr `yaml:"httpListenAddress"`
	GRPCListenAddress *pkgnet.ListenAddr `yaml:"grpcListenAddress"`
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts chat server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	var (
		logger   = p.Logger
		registry = prometheus.NewRegistry()
	)

	httpListener, err := p.HTTPListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to http address: %w", err)
	}

	grpcListener, err := p.GRPCListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("binding to gRPC address: %w", err)
	}

	cs, err := NewEventServer(logger, registry)
	if err != nil {
		return fmt.Errorf("constructing chat server: %w", err)
	}

	r, err := routes(logger, registry, cs)
	if err != nil {
		return fmt.Errorf("configuring routes: %v", err)
	}

	var (
		httpSrv    = pkghttp.NewServer(logger, httpListener, r)
		grpcServer = grpc.NewServer()

		errCh = make(chan error, 2)
	)

	go func() {
		logger.Info("starting http server", zap.Stringer("address", p.HTTPListenAddress))

		errCh <- httpSrv.Start()
	}()

	go func() {
		eventserverpb.RegisterEventServerServer(grpcServer, NewGRPCServer(logger, cs))

		logger.Info("starting gRPC server", zap.Stringer("address", p.GRPCListenAddress))

		errCh <- grpcServer.Serve(grpcListener)
	}()

	select {
	case <-p.CloseCh:
		logger.Info("received close request, terminating")
		if err = httpSrv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shuting down server: %w", err)
		}

		// TODO: implement graceful shutdown for gRPC server
	case err = <-errCh:
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}

func routes(l *zap.Logger, r *prometheus.Registry, server EventServer) (http.Handler, error) {
	rawRouter := pkghttp.NewRouterBuilder(l)
	rawRouter.HandleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP).
		Methods(http.MethodGet)

	instrumentedRouter := rawRouter.SubGroup().
		With(httpmiddleware.BasicMetrics(r), httpmiddleware.RequestLogger(l))

	instrumentedRouter.HandleJSONFunc("/health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodGet)

	// TODO: protobuf alternative should be added too
	instrumentedRouter.HandleJSONFunc("/send-message", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var msg Event
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			return nil, err
		}

		if err := server.SendEvent(r.Context(), msg); err != nil {
			return nil, fmt.Errorf("sending message: %w", err)
		}

		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	return rawRouter.Build(), nil
}
