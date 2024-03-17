package eventserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcmeta "google.golang.org/grpc/metadata"

	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	"github.com/faustuzas/occa/src/pkg/generated/proto/rteventspb"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

type GRPCTester struct {
	t      *testing.T
	ctx    context.Context
	client eventserverpb.EventServerClient

	userID pkgid.ID
	token  string
}

func NewGRPCTester(t *testing.T, ctx context.Context, address string, userID pkgid.ID, name string) *GRPCTester {
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &GRPCTester{
		t:      t,
		ctx:    ctx,
		userID: userID,
		token:  generateToken(t, userID, name),
		client: eventserverpb.NewEventServerClient(conn),
	}
}

func (t *GRPCTester) Connect() GRPCStream[*rteventspb.Event] {
	md := grpcmeta.Pairs("authorization", fmt.Sprintf("Bearer %v", t.token))
	ctx := grpcmeta.NewOutgoingContext(t.ctx, md)

	stream, err := t.client.Connect(ctx, &eventserverpb.ConnectRequest{
		UserId: t.userID.String(),
	})
	require.NoError(t.t, err)

	return startGRPCReceiver[*rteventspb.Event](stream)
}

func (t *GRPCTester) RawGRPCClient() eventserverpb.EventServerClient {
	return t.client
}

type Stream[T any] interface {
	Recv() (T, error)
}

type Msg[T any] struct {
	El  T
	Err error
}

type GRPCStream[T any] struct {
	stream Stream[T]

	outputCh chan Msg[T]
}

func startGRPCReceiver[T any](stream Stream[T]) GRPCStream[T] {
	ch := make(chan Msg[T])

	r := GRPCStream[T]{
		stream:   stream,
		outputCh: ch,
	}

	go func() {
		el, err := stream.Recv()
		ch <- Msg[T]{
			El:  el,
			Err: err,
		}
	}()

	return r
}

func (r GRPCStream[T]) RecvCh() <-chan Msg[T] {
	return r.outputCh
}

func generateToken(t *testing.T, userID pkgid.ID, name string) string {
	_, privatePath, err := pkgtest.GetRSAPairPaths()
	require.NoError(t, err)

	privateKey, err := pkgauth.ReadPrivateKey(privatePath)
	require.NoError(t, err)

	issuer := pkgauth.NewJWTIssuer(privateKey, time.Now)

	token, err := issuer.Issue(context.Background(), pkgauth.Principal{
		ID:       userID,
		UserName: name,
	})
	return token
}
