package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	pkgnet "github.com/faustuzas/occa/src/pkg/net"
)

type Server struct {
	s       *http.Server
	address *pkgnet.ListenAddr
	logger  *zap.Logger
}

func NewServer(logger *zap.Logger, address *pkgnet.ListenAddr, handler http.Handler) Server {
	return Server{
		s: &http.Server{
			Handler:      handler,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
		address: address,
		logger:  logger,
	}
}

func (s Server) Start() error {
	l, err := s.address.Listener()
	if err != nil {
		return fmt.Errorf("acquiring listener: %w", err)
	}

	err = s.s.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s Server) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
