package serviceboot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway"
	"github.com/faustuzas/occa/src/integration/containers"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

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
			HTTPListenAddress: listenAddr,

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
		Logger:  pkgtest.Logger.With(zap.String("component", "gateway")),
		CloseCh: closeCh,
	}
}
