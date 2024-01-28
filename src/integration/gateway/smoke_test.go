package gateway

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgauthdb "github.com/faustuzas/occa/src/pkg/auth/db"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkginmemorydb "github.com/faustuzas/occa/src/pkg/inmemorydb"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgservice "github.com/faustuzas/occa/src/pkg/service"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGatewaySmoke(t *testing.T) {
	var (
		ctrl = gomock.NewController(t)

		usersDB = pkgauthdb.NewMockUsers(ctrl)
		store   = pkginmemorydb.NewMockStore(ctrl)
	)

	usersDB.EXPECT().Start().Return(nil).AnyTimes()
	usersDB.EXPECT().Close().Return(nil).AnyTimes()

	store.EXPECT().Close().Return(nil).AnyTimes()

	closeCh := make(chan struct{})
	defer func() {
		close(closeCh)
	}()

	listenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")

	// bound to the address so the port would be allocated now
	_, err := listenAddr.Listener()
	require.NoError(t, err)

	go func() {
		err := gateway.Start(gateway.Params{
			Configuration: gateway.Configuration{
				ServerListenAddress: listenAddr,

				InMemoryDB: pkgservice.FromImpl[pkginmemorydb.Store, pkginmemorydb.Configuration](store),
				Auth:       pkgauth.ValidatorConfiguration{Type: pkgauth.ValidatorConfigurationNoop},
				Registerer: pkgauth.RegistererConfiguration{
					UsersDB:     pkgservice.FromImpl[pkgauthdb.Users, pkgauth.UsersConfiguration](usersDB),
					TokenIssuer: pkgservice.FromImpl[pkgauth.TokenIssuer, pkgauth.TokenIssuerConfiguration](pkgauth.NewMockTokenIssuer(ctrl)),
				},
			},
			Logger:   pkgtest.Logger,
			Registry: prometheus.NewRegistry(),
			CloseCh:  closeCh,
		})

		require.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	_, body := pkgtest.HTTPGetBody(t, listenAddr.String(), "/health")
	require.Equal(t, pkghttp.DefaultOKResponse(), string(body))
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
