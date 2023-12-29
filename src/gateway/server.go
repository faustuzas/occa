package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	gatewayclient "github.com/faustuzas/occa/src/gateway/client"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkginmemorydb "github.com/faustuzas/occa/src/pkg/inmemorydb"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgservice "github.com/faustuzas/occa/src/pkg/service"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	ServerListenAddress *pkgnet.ListenAddr `yaml:"listenAddress"`

	InMemoryDB *pkgservice.ExternalService[pkginmemorydb.Store, pkginmemorydb.Configuration] `yaml:"inMemoryDB"`
	Auth       pkgauth.ValidatorConfiguration                                                `yaml:"auth"`
	Registerer pkgauth.RegistererConfiguration                                               `yaml:"registerer"`
}

type Params struct {
	Configuration

	Logger   *zap.Logger
	Registry *prometheus.Registry

	CloseCh <-chan struct{}
}

// Start starts gateway server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	logger := p.Logger

	listener, err := p.ServerListenAddress.Listener()
	if err != nil {
		return fmt.Errorf("unable to listen: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	services, err := p.StartServices()
	if err != nil {
		return err
	}
	defer func() {
		if e := services.Close(); e != nil {
			logger.Error("error while closing services", zap.Error(e))
		}
	}()

	routes, err := configureRoutes(p, services)
	if err != nil {
		return fmt.Errorf("configuring routes: %v", err)
	}

	var (
		srv = http.Server{
			Handler:      routes,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		}
		srvErrCh = make(chan error, 1)
	)
	go func() {
		logger.Info("starting server", zap.Stringer("address", p.ServerListenAddress))

		err = srv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			srvErrCh <- err
		}
	}()

	select {
	case <-p.CloseCh:
		logger.Info("received close request, terminating")
		if err = srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("closing server: %w", err)
		}
	case err = <-srvErrCh:
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}

func configureRoutes(p Params, services Services) (http.Handler, error) {
	rawRouter := pkghttp.NewRouterBuilder(p.Logger)
	rawRouter.HandleFunc("/metrics", promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{}).ServeHTTP).
		Methods(http.MethodGet)

	instrumentedRouter := rawRouter.SubGroup().
		With(httpmiddleware.BasicMetrics(p.Registry), httpmiddleware.RequestLogger(p.Logger))

	instrumentedRouter.HandleJSONFunc("/health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodGet)

	instrumentedRouter.HandleJSONFunc("/register", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req gatewayclient.RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		err := services.AuthRegisterer.Register(req.Username, req.Password)
		if err != nil {
			return nil, err
		}

		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	instrumentedRouter.HandleJSONFunc("/login", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req gatewayclient.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		token, err := services.AuthRegisterer.Login(req.Username, req.Password)
		if err != nil {
			return nil, err
		}

		return gatewayclient.LoginResponse{
			Token: token,
		}, nil
	}).Methods(http.MethodPost)

	authenticatedRouter := instrumentedRouter.SubGroup().
		With(services.HTTPAuthMiddleware)

	authenticatedRouter.HandleJSONFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) (any, error) {
		principal := pkgauth.PrincipalFromContext(r.Context())

		if err := services.InMemoryDB.SetCollectionItemWithTTL(
			r.Context(), "active_users",
			principal.UserName,
			"",
			30*time.Second,
		); err != nil {
			return nil, fmt.Errorf("writing into redis: %w", err)
		}
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	authenticatedRouter.HandleJSONFunc("/active-users", func(w http.ResponseWriter, r *http.Request) (any, error) {
		activeUsers, err := services.InMemoryDB.ListCollectionKeys(r.Context(), "active_users")
		if err != nil {
			return nil, fmt.Errorf("was not to read from redis: %w", err)
		}

		return gatewayclient.ActiveUsersResponse{
			ActiveUsers: activeUsers,
		}, nil
	}).Methods(http.MethodGet)

	return rawRouter.Build(), nil
}
