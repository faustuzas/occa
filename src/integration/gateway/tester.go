package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	gatewayhttp "github.com/faustuzas/occa/src/gateway/http"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

type Tester struct {
	t       *testing.T
	ctx     context.Context
	address string

	token string
}

func NewTester(t *testing.T, ctx context.Context, address string) *Tester {
	return &Tester{t: t, ctx: ctx, address: address}
}

func (g *Tester) Register(username, password string) {
	req := g.createReq(http.MethodPost, "/register", pkgtest.ToJSONBody(g.t, gatewayhttp.RegistrationRequest{
		Username: username,
		Password: password,
	}))

	resp, _ := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)
}

func (g *Tester) Login(username, password string) string {
	req := g.createReq(http.MethodPost, "/login", pkgtest.ToJSONBody(g.t, gatewayhttp.LoginRequest{
		Username: username,
		Password: password,
	}))

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	loginResp := pkgtest.FromJSONBytes[gatewayhttp.LoginResponse](g.t, body)
	require.NotEmpty(g.t, loginResp.Token)

	g.token = loginResp.Token
	return loginResp.Token
}

func (g *Tester) Heartbeat() {
	req := g.createAuthenticatedReq(http.MethodPost, "/heartbeat", nil)

	resp, _ := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)
}

func (g *Tester) ActiveUsers() gatewayhttp.ActiveUsersResponse {
	req := g.createAuthenticatedReq(http.MethodGet, "/active-users", nil)

	resp, body := pkgtest.HTTPExec(g.t, req)
	require.Equal(g.t, http.StatusOK, resp.StatusCode)

	return pkgtest.FromJSONBytes[gatewayhttp.ActiveUsersResponse](g.t, body)
}

func (g *Tester) createReq(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(g.ctx, method, "http://"+g.address+path, body)
	require.NoError(g.t, err)

	return req
}

func (g *Tester) createAuthenticatedReq(method string, path string, body io.Reader) *http.Request {
	require.NotEmpty(g.t, g.token)

	req := g.createReq(method, path, body)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))

	return req
}
