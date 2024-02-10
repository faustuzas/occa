package eventserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/faustuzas/occa/src/eventserver"
	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"

	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestChatServer_ClientConnect_gRPC(t *testing.T) {
	params := serviceboot.DefaultEventServerParams(t)
	go func() {
		require.NoError(t, eventserver.Start(params))
	}()

	ctx := context.Background()

	conn, err := grpc.DialContext(
		ctx,
		params.GRPCListenAddress.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)

	client := eventserverpb.NewEventServerClient(conn)

	recipientID := pkgid.NewID()
	stream, err := client.Connect(ctx, &eventserverpb.ConnectRequest{UserId: recipientID.String()})
	require.NoError(t, err)

	wait := make(chan struct{})
	go func() {
		defer close(wait)

		var m eventserverpb.Event
		for {
			require.NoError(t, stream.RecvMsg(&m))
			fmt.Printf("Received: %v\n", m.String())
			return
		}
	}()

	csTester := newChatServerTester(t, ctx, params.HTTPListenAddress.String())
	csTester.sendMessage(pkgid.NewID(), recipientID, "hello world!")

	<-wait
}

func newChatServerTester(t *testing.T, ctx context.Context, httpAddress string) *chatServerTester {
	return &chatServerTester{t: t, ctx: ctx, httpAddress: httpAddress}
}

type chatServerTester struct {
	t           *testing.T
	ctx         context.Context
	httpAddress string
}

func (t *chatServerTester) sendMessage(from pkgid.ID, to pkgid.ID, message string) {
	req := t.createReq(http.MethodPost, "/send-message", toBody(t.t, eventserver.Event{
		SenderID:         from,
		RecipientID:      to,
		Content:          message,
		SentFromClientAt: time.Now(),
	}))

	resp, _ := pkgtest.HTTPExec(t.t, req)
	require.Equal(t.t, http.StatusOK, resp.StatusCode)
}

func (t *chatServerTester) createReq(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(t.ctx, method, "http://"+t.httpAddress+path, body)
	require.NoError(t.t, err)

	return req
}

func toBody(t *testing.T, val any) io.Reader {
	b, err := json.Marshal(val)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
