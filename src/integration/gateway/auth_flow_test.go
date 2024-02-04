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
	gatewayhttp "github.com/faustuzas/occa/src/gateway/http"
	"github.com/faustuzas/occa/src/integration/containers"
	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGateway_AuthFlow(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	gatewayParams := serviceboot.DefaultGatewayParams(t, containers.WithMysql(t), containers.WithRedis(t))
	go func() {
		err := gateway.Start(gatewayParams)
		cancelFn()
		require.NoError(t, err)
	}()

	gTester := gatewayTester{t: t, ctx: ctx, address: gatewayParams.ServerListenAddress.String()}

	// Perform authorized call and fail.
	req := gTester.createReq(http.MethodPost, "/heartbeat", nil)
	resp, _ := pkgtest.HTTPExec(t, req)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	gTester.register("user", "password")
	token := gTester.login("user", "password")

	// Perform authorized call and succeed.
	req = gTester.createReq(http.MethodPost, "/heartbeat", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = pkgtest.HTTPExec(t, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func newGatewayTester(t *testing.T, ctx context.Context, address string) *gatewayTester {
	return &gatewayTester{t: t, ctx: ctx, address: address}
}

type gatewayTester struct {
	t       *testing.T
	ctx     context.Context
	address string

	token string
}

func (g *gatewayTester) register(username, password string) {
	req := g.createReq(http.MethodPost, "/register", toBody(g.t, gatewayhttp.RegistrationRequest{
		Username: username,
		Password: password,
	}))

	resp, _ := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)
}

func (g *gatewayTester) login(username, password string) string {
	req := g.createReq(http.MethodPost, "/login", toBody(g.t, gatewayhttp.LoginRequest{
		Username: username,
		Password: password,
	}))

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	var loginResp gatewayhttp.LoginResponse
	require.NoError(g.t, json.Unmarshal(body, &loginResp))
	require.NotEmpty(g.t, loginResp.Token)

	g.token = loginResp.Token
	return loginResp.Token
}

func (g *gatewayTester) createReq(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(g.ctx, method, "http://"+g.address+path, body)
	require.NoError(g.t, err)

	return req
}

func (g *gatewayTester) createAuthenticatedReq(method string, path string, body io.Reader) *http.Request {
	require.NotEmpty(g.t, g.token)

	req := g.createReq(method, path, body)
	req.Header.Add("Authorization", g.token)
	return req
}

func toBody(t *testing.T, val any) io.Reader {
	b, err := json.Marshal(val)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
