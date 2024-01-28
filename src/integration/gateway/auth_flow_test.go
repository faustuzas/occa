package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	gatewayclient "github.com/faustuzas/occa/src/gateway/client"
	"github.com/faustuzas/occa/src/integration/containers"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestAuthFlowE2E(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	gatewayParams := DefaultGatewayParams(t, containers.WithMysql(t), containers.WithRedis(t))
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

func DefaultGatewayParams(t *testing.T, db *containers.MySQLContainer, redis *containers.RedisContainer) gateway.Params {
	listenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")
	listener, err := listenAddr.Listener()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
	})

	closeCh := make(chan struct{})
	t.Cleanup(func() {
		close(closeCh)
	})

	authDatabase, err := db.WithTemporaryDatabase(t, "auth_users")
	require.NoError(t, err)

	pubKey, privKey, err := pkgtest.GetRSAPairPaths()
	require.NoError(t, err)

	return gateway.Params{
		Configuration: gateway.Configuration{
			ServerListenAddress: listenAddr,

			MemStore: pkgmemstore.Configuration{
				User:     redis.Username,
				Password: redis.Password,
				Address:  fmt.Sprintf("localhost:%d", redis.Port),
			},

			Auth: pkgauth.ValidatorConfiguration{
				Type: pkgauth.ValidatorConfigurationJWTRSA,
				JWTValidator: pkgauth.JWTValidatorConfiguration{
					PublicKeyPath: pubKey,
				},
			},

			Registerer: pkgauth.RegistererConfiguration{
				TokenIssuer: pkgauth.TokenIssuerConfiguration{
					PrivateKeyPath: privKey,
				},
				Users: pkgauth.UsersConfiguration{
					DB: pkgdb.Configuration{
						DBType:         "mysql",
						DataSourceName: db.DataSourceName(authDatabase),
					},
				},
			},
		},
		Logger:   pkgtest.Logger,
		Registry: prometheus.NewRegistry(),
		CloseCh:  closeCh,
	}
}
