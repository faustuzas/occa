package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	s      *http.Server
	l      net.Listener
	logger *zap.Logger
}

func NewServer(logger *zap.Logger, l net.Listener, handler http.Handler) Server {
	return Server{
		s: &http.Server{
			Handler:      handler,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
		l:      l,
		logger: logger,
	}
}

func (s Server) Start() error {
	if err := s.s.Serve(s.l); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s Server) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
