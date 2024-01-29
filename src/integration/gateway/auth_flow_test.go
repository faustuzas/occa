package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	gatewayclient "github.com/faustuzas/occa/src/gateway/client"
	"github.com/faustuzas/occa/src/integration/containers"
	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestAuthFlowE2E(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	gatewayParams := serviceboot.DefaultGatewayParams(t, containers.WithMysql(t), containers.WithRedis(t))
	go func() {
		require.NoError(t, gateway.Start(gatewayParams))
	}()

	gTester := gatewayTester{t: t, ctx: ctx, address: gatewayParams.ServerListenAddress.String()}

	// Registration step.
	{
		req := gTester.createReq(http.MethodPost, "/register", toBody(t, gatewayclient.RegistrationRequest{
			Username: "user",
			Password: "password",
		}))

		resp, _ := pkgtest.HTTPExec(t, req)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Login step.
	var token string
	{
		req := gTester.createReq(http.MethodPost, "/login", toBody(t, gatewayclient.LoginRequest{
			Username: "user",
			Password: "password",
		}))

		resp, body := pkgtest.HTTPExec(t, req)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp gatewayclient.LoginResponse
		require.NoError(t, json.Unmarshal(body, &loginResp))
		require.NotEmpty(t, loginResp.Token)

		token = loginResp.Token
	}

	// Perform authorized call.
	{
		req := gTester.createReq(http.MethodPost, "/heartbeat", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := pkgtest.HTTPExec(t, req)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

type gatewayTester struct {
	t       *testing.T
	ctx     context.Context
	address string
}

func (g *gatewayTester) createReq(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(g.ctx, method, "http://"+g.address+path, body)
	require.NoError(g.t, err)

	return req
}

func toBody(t *testing.T, val any) io.Reader {
	b, err := json.Marshal(val)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
