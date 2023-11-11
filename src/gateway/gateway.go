package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	pkgserver "github.com/faustuzas/tcha/src/pkg/server"
)

type Configuration struct {
	ServerListenAddress *pkgserver.ListenAddr
}

type Params struct {
	Configuration

	Logger *zap.Logger

	CloseCh <-chan struct{}
}

// Start starts gateway server and blocks until close request is received.
// Returns error in case initialisation failed.
func Start(p Params) error {
	logger := p.Logger.With(zap.String("component", "gateway"))

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
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok")) // add some proper error handling
	}).Methods(http.MethodGet)

	return r, nil
}
