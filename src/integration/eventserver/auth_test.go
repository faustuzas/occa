package eventserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/faustuzas/occa/src/eventserver"
	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

func TestEventServer_GRPCAuth(t *testing.T) {
	var (
		ctx    = context.Background()
		params = DefaultParams(t)

		userID = pkgid.NewID()
	)
	go func() {
		require.NoError(t, eventserver.Start(params))
	}()

	rawConn, err := grpc.DialContext(ctx, params.GRPCListenAddress.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() {
		_ = rawConn.Close()
	}()

	stream, err := eventserverpb.NewEventServerClient(rawConn).Connect(ctx, &eventserverpb.ConnectRequest{
		UserId: userID.String(),
	})
	require.NoError(t, err) // establishing the stream does not immediately invoke the auth middleware

	_, err = stream.Recv()
	require.Error(t, err, "missing Authorization header")
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	tester := NewGRPCTester(t, ctx, params.GRPCListenAddress.String(), userID, "test1")
	select {
	case msg := <-tester.Connect().RecvCh():
		require.Failf(t, "was not supposed to receive anything", "msg: %+v", msg)
	case <-time.After(20 * time.Millisecond):
	}
}
