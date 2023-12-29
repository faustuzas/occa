package gateway

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/faustuzas/occa/src/gateway"
	"github.com/faustuzas/occa/src/integration/helpers"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgredis "github.com/faustuzas/occa/src/pkg/redis"
	pkgservice "github.com/faustuzas/occa/src/pkg/service"
)

func TestGatewaySmoke(t *testing.T) {
	ctrl := gomock.NewController(t)

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

				Redis: pkgservice.FromImplementation[pkgredis.Client, pkgredis.Configuration](pkgredis.NewMockClient(ctrl)),
			},
			Logger:   helpers.Logger,
			Registry: prometheus.NewRegistry(),
			CloseCh:  closeCh,
		})

		require.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	_, body := helpers.HTTPGetBody(t, listenAddr.String(), "/health")
	require.Equal(t, pkghttp.DefaultOKResponse(), string(body))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
