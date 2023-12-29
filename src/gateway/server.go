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
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgredis "github.com/faustuzas/occa/src/pkg/redis"
	pkgservice "github.com/faustuzas/occa/src/pkg/service"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	ServerListenAddress *pkgnet.ListenAddr `yaml:"listenAddress"`

	Redis *pkgservice.ExternalService[pkgredis.Client, pkgredis.Configuration] `yaml:"redis"`
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

	routes, err := configureRoutes(p)
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

func configureRoutes(p Params) (http.Handler, error) {
	redisClient, err := p.Configuration.Redis.GetService()
	if err != nil {
		return nil, fmt.Errorf("building redis client: %w", err)
	}

	r := pkghttp.NewRouterBuilder(p.Logger)
	r.HandleFunc("/metrics", promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{}).ServeHTTP).
		Methods(http.MethodGet)

	serviceMiddlewares := []httpmiddleware.Middleware{
		httpmiddleware.BasicMetrics(p.Registry),
		httpmiddleware.RequestLogger(p.Logger),
		pkgauth.HTTPMiddleware(p.Logger),
	}

	serviceRouter := r.SubGroup().With(serviceMiddlewares...)

	serviceRouter.HandleJSONFunc("/health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodGet)

	serviceRouter.HandleJSONFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req gatewayclient.AuthenticationRequest
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		if err = redisClient.PutIntoCollectionWithTTL(
			r.Context(), "active_users",
			req.Username,
			"",
			30*time.Second,
		); err != nil {
			return nil, err
		}

		return gatewayclient.AuthenticationResponse{
			Token: fmt.Sprintf("token-%v", req.Username),
		}, nil
	}).Methods(http.MethodPost)

	serviceRouter.HandleJSONFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) (any, error) {
		token := r.Header.Get("Authorization")
		if token == "" {
			return nil, pkghttp.ErrUnauthorized(fmt.Errorf("missing token"))
		}

		username := token[len("token-"):]
		if err := redisClient.PutIntoCollectionWithTTL(
			r.Context(), "active_users",
			username,
			"",
			30*time.Second,
		); err != nil {
			return nil, fmt.Errorf("writing into redis: %w", err)
		}
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	serviceRouter.HandleJSONFunc("/active-users", func(w http.ResponseWriter, r *http.Request) (any, error) {
		activeUsers, err := redisClient.ListCollection(r.Context(), "active_users")
		if err != nil {
			return nil, fmt.Errorf("was not to read from redis: %w", err)
		}

		users := make([]string, 0, len(activeUsers))
		for usr := range activeUsers {
			users = append(users, usr)
		}

		return gatewayclient.ActiveUsersResponse{
			ActiveUsers: users,
		}, nil
	}).Methods(http.MethodGet)

	return r.Build(), nil
}
