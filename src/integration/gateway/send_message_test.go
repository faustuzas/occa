package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/faustuzas/occa/src/eventserver/generated/proto/eventserverpb"
	gatewayhttp "github.com/faustuzas/occa/src/gateway/http"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func Test(t *testing.T) {
	var (
		ctx = context.Background()

		gatewayAddress = "localhost:9000"

		userName = "faustas:" + strconv.Itoa(rand.Intn(10000))
		pass     = "pass"
	)

	gTester := newGatewayTester(t, ctx, gatewayAddress)

	gTester.register(userName, pass)
	gTester.login(userName, pass)

	gTester.heartbeat()
	users := gTester.activeUsers().ActiveUsers
	require.NotEmpty(t, users)

	myID := users[0].ID

	selectedServer := gTester.selectServer()
	fmt.Printf("Selected server: %v\n", selectedServer)

	conn, err := grpc.DialContext(
		ctx,
		selectedServer.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	require.NoError(t, err)

	client := eventserverpb.NewEventServerClient(conn)

	stream, err := client.Connect(ctx, &eventserverpb.ConnectRequest{UserId: myID.String()})
	require.NoError(t, err)

	go func() {
		var m eventserverpb.Event
		for {
			require.NoError(t, stream.RecvMsg(&m))
			fmt.Printf("Received: %v\n", m.String())
		}
	}()

	for i := 0; i < 100; i++ {
		gTester.sendMessage(users[0].ID, fmt.Sprintf("hello there: %d", i))
		time.Sleep(5 * time.Second)
	}
}

func (g *gatewayTester) selectServer() gatewayhttp.SelectServerResponse {
	req := g.createAuthenticatedReq(http.MethodGet, "/select-server", nil)

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	var r gatewayhttp.SelectServerResponse
	require.NoError(g.t, json.Unmarshal(body, &r))

	return r
}

func (g *gatewayTester) sendMessage(recipientID pkgid.ID, message string) gatewayhttp.SelectServerResponse {
	req := g.createAuthenticatedReq(http.MethodPost, "/send-message", toBody(g.t, gatewayhttp.SendMessageRequest{
		RecipientID: recipientID,
		Message:     message,
	}))

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	var r gatewayhttp.SelectServerResponse
	require.NoError(g.t, json.Unmarshal(body, &r))

	return r
}
