package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	gatewayclient "github.com/faustuzas/occa/src/gateway/client"
	pkgconfig "github.com/faustuzas/occa/src/pkg/config"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgredis "github.com/faustuzas/occa/src/pkg/redis"
)

type Configuration struct {
	pkgconfig.CommonConfiguration `yaml:",inline"`

	ServerListenAddress *pkgnet.ListenAddr     `yaml:"listenAddress"`
	Redis               pkgredis.Configuration `yaml:"redis"`
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts gateway server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	logger := p.Logger

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
		listener, err := p.ServerListenAddress.Listener()
		if err != nil {
			srvErrCh <- err
			return
		}

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
	redisClient, err := p.Configuration.Redis.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("building redis client: %w", err)
	}

	r := pkghttp.NewRouter(p.Logger)
	r.Use(httpmiddleware.RequestLogger(p.Logger))

	r.HandleJSONFunc("/health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return "ok", nil
	}).Methods(http.MethodGet)

	r.HandleJSONFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) (any, error) {
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

	r.HandleJSONFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) (any, error) {
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
		return "ok", nil
	}).Methods(http.MethodPost)

	r.HandleJSONFunc("/active-users", func(w http.ResponseWriter, r *http.Request) (any, error) {
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

	return r, nil
}
