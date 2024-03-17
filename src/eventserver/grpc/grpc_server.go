package grpc

import (
	"context"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"
	"github.com/faustuzas/occa/src/eventserver/services"
	"github.com/faustuzas/occa/src/pkg/generated/proto/rteventspb"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

var _ eventserverpb.EventServerServer = (*Server)(nil)

type Services struct {
	EventServer           services.EventServer
	StreamAuthInterceptor grpc.StreamServerInterceptor

	Instrumentation pkginstrument.Instrumentation
}

func Configure(services Services) (*grpc.Server, error) {
	metrics := grpcprom.NewServerMetrics()
	services.Instrumentation.Registerer.MustRegister(metrics)

	grpcServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			metrics.StreamServerInterceptor(),
			services.StreamAuthInterceptor,
		),
	)

	server := NewServer(services.Instrumentation, services.EventServer)
	eventserverpb.RegisterEventServerServer(grpcServer, server)

	return grpcServer, nil
}

func NewServer(i pkginstrument.Instrumentation, eventServer services.EventServer) *Server {
	return &Server{
		eventServer: eventServer,

		i: i,
	}
}

type Server struct {
	eventserverpb.UnimplementedEventServerServer

	eventServer services.EventServer
	i           pkginstrument.Instrumentation
}

func (s *Server) Connect(req *eventserverpb.ConnectRequest, server eventserverpb.EventServer_ConnectServer) error {
	id := pkgid.FromString(req.UserId)

	s.i.Logger.Info("new gRPC-based user connected", zap.Stringer("id", id))

	return s.eventServer.ServeConnection(id, NewGRPCConnection(server))
}

type grpcConnection struct {
	sender eventserverpb.EventServer_ConnectServer
}

func NewGRPCConnection(sender eventserverpb.EventServer_ConnectServer) services.Connection {
	return &grpcConnection{
		sender: sender,
	}
}

func (g grpcConnection) SendEvent(_ context.Context, msg services.Event) error {
	return g.sender.Send(&rteventspb.Event{
		Payload: &rteventspb.Event_DirectMessage{
			DirectMessage: &rteventspb.DirectMessage{
				SenderId: msg.SenderID.String(),
				Message:  msg.Content,
			},
		},
	})
}
