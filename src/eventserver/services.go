package eventserver

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/faustuzas/occa/src/eventserver/services"
	"github.com/faustuzas/occa/src/pkg/eventserver/membership"
	"github.com/faustuzas/occa/src/pkg/eventserver/rtconn"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgio "github.com/faustuzas/occa/src/pkg/io"
)

type Services struct {
	pkgio.Closers

	HTTPAuthMiddleware        httpmiddleware.Middleware
	GRPCStreamAuthInterceptor grpc.StreamServerInterceptor

	EventServer       services.EventServer
	MembershipManager membership.Manager

	MetricsRegistry *prometheus.Registry
}

func (p Params) StartServices() (Services, error) {
	var (
		registry = prometheus.NewRegistry()

		inst = pkginstrument.Instrumentation{
			Logger:     p.Logger,
			Registerer: registry,
		}

		closers pkgio.Closers
	)

	etcdClient, err := p.Configuration.Etcd.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building etcd client: %w", err)
	}
	closers = append(closers, pkgio.CloseWithoutContext(etcdClient.Close))

	memstore, err := p.MemStore.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building memstore client: %w", err)
	}
	closers = append(closers, pkgio.CloseWithoutContext(memstore.Close))

	hearthBeater := rtconn.NewHeartBeater(inst, p.ServerID, memstore)
	closers = append(closers, hearthBeater)

	eventServer, err := services.NewEventServer(inst, hearthBeater)
	if err != nil {
		return Services{}, fmt.Errorf("building events server: %w", err)
	}

	httpAuthMiddleware, err := p.Configuration.Auth.BuildHTTPMiddleware(inst)
	if err != nil {
		return Services{}, fmt.Errorf("building HTTP auth middleware: %w", err)
	}

	grpcAuthMiddleware, err := p.Configuration.Auth.BuildGRPCStreamInterceptor(inst)
	if err != nil {
		return Services{}, fmt.Errorf("building gRPC auth middleware: %w", err)
	}

	membershipManager, err := membership.NewManager(inst, etcdClient)
	if err != nil {
		return Services{}, fmt.Errorf("creating membership manager: %w", err)
	}

	return Services{
		HTTPAuthMiddleware:        httpAuthMiddleware,
		GRPCStreamAuthInterceptor: grpcAuthMiddleware,
		EventServer:               eventServer,
		MembershipManager:         membershipManager,
		MetricsRegistry:           registry,
		Closers:                   closers,
	}, nil
}
