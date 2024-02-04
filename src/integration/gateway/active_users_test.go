package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	gatewayhttp "github.com/faustuzas/occa/src/gateway/http"
	"github.com/faustuzas/occa/src/integration/containers"
	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGateway_ActiveUsers(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	gatewayParams := serviceboot.DefaultGatewayParams(t, containers.WithMysql(t), containers.WithRedis(t))
	go func() {
		err := gateway.Start(gatewayParams)
		cancelFn()
		require.NoError(t, err)
	}()

	user1G := newGatewayTester(t, ctx, gatewayParams.HTTPListenAddress.String())
	user1G.register("user_1", "password")
	user1G.login("user_1", "password")

	activeUsers := user1G.activeUsers().ActiveUsers
	require.Empty(t, activeUsers)

	user1G.heartbeat()

	activeUsers = user1G.activeUsers().ActiveUsers
	require.Len(t, activeUsers, 1)
	require.Equal(t, "user_1", activeUsers[0].Username)
	require.NotEmpty(t, activeUsers[0].ID)
	require.NotEmpty(t, activeUsers[0].LastSeen)
}

func (g *gatewayTester) heartbeat() {
	req := g.createAuthenticatedReq(http.MethodPost, "/heartbeat", nil)

	resp, _ := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)
}

func (g *gatewayTester) activeUsers() gatewayhttp.ActiveUsersResponse {
	req := g.createAuthenticatedReq(http.MethodGet, "/active-users", nil)

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	var r gatewayhttp.ActiveUsersResponse
	require.NoError(g.t, json.Unmarshal(body, &r))

	return r
}
