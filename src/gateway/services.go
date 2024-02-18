package gateway

import (
	"context"
	"fmt"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/faustuzas/occa/src/gateway/services"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgclock "github.com/faustuzas/occa/src/pkg/clock"
	esclient "github.com/faustuzas/occa/src/pkg/eventserver/client"
	esmembership "github.com/faustuzas/occa/src/pkg/eventserver/membership"
	"github.com/faustuzas/occa/src/pkg/eventserver/rtconn"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
	pkgio "github.com/faustuzas/occa/src/pkg/io"
)

type Services struct {
	pkgio.Closers

	HTTPAuthMiddleware  httpmiddleware.Middleware
	AuthRegisterer      pkgauth.Registerer
	ActiveUserTracker   services.ActiveUsersTracker
	RTEventsRelay       services.RealTimeEventRelay
	EventServerRegistry *esmembership.ServerRegistry

	MetricsRegistry *prometheus.Registry
}

func (p Params) StartServices() (_ Services, err error) {
	var (
		registry = prometheus.NewRegistry()

		inst = pkginstrument.Instrumentation{
			Logger:     p.Logger,
			Registerer: registry,
		}

		starters pkgio.Starters
		closers  pkgio.Closers

		clock = pkgclock.RealClock{}
	)
	defer func() {
		if err != nil {
			err = multierr.Append(err, closers.Close(context.Background()))
		}
	}()

	memStore, err := p.Configuration.MemStore.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building redis client: %w", err)
	}
	closers = append(closers, pkgio.CloseWithoutContext(memStore.Close))

	httpAuthMiddleware, err := p.Configuration.Auth.BuildHTTPMiddleware(inst)
	if err != nil {
		return Services{}, fmt.Errorf("building HTTP auth middleware: %w", err)
	}

	tokenIssuer, err := p.Registerer.TokenIssuer.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building JWT token issuer: %w", err)
	}

	usersDB, err := p.Registerer.Users.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building auth db connection: %w", err)
	}
	starters = append(starters, usersDB)
	closers = append(closers, usersDB)

	activeUsersTracker, err := services.NewActiveUsersTracker(memStore, clock)
	if err != nil {
		return Services{}, fmt.Errorf("building active users tracker: %w", err)
	}

	etcdClient, err := p.Etcd.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building etcd client: %w", err)
	}

	eventServersRegistry := esmembership.NewServerRegistry(inst, etcdClient)
	starters = append(starters, eventServersRegistry)
	closers = append(closers, eventServersRegistry)

	esPool := esclient.NewPool(inst, eventServersRegistry)
	closers = append(closers, esPool)

	rtServerResolver := rtconn.NewServerResolver(inst, memStore)
	rtRelay := services.NewRealTimeEventRelay(inst, rtServerResolver, esPool)

	if err = starters.Start(context.Background()); err != nil {
		return Services{}, fmt.Errorf("starting services: %w", err)
	}

	return Services{
		ActiveUserTracker:   activeUsersTracker,
		HTTPAuthMiddleware:  httpAuthMiddleware,
		AuthRegisterer:      pkgauth.NewRegisterer(usersDB, tokenIssuer),
		RTEventsRelay:       rtRelay,
		EventServerRegistry: eventServersRegistry,
		MetricsRegistry:     registry,

		Closers: closers,
	}, nil
}
