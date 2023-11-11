package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/faustuzas/tcha/src/gateway"
	"github.com/faustuzas/tcha/src/integration/helpers"
	pkgserver "github.com/faustuzas/tcha/src/pkg/server"
)

func TestGatewaySmoke(t *testing.T) {
	closeCh := make(chan struct{})
	defer func() {
		close(closeCh)
	}()

	listenAddr := pkgserver.ListenAddrFromAddress("0.0.0.0:0")

	// bound to the address so the port would be allocated now
	_, err := listenAddr.Listener()
	require.NoError(t, err)

	go func() {
		err := gateway.Start(gateway.Params{
			Configuration: gateway.Configuration{
				ServerListenAddress: listenAddr,
			},
			Logger:  helpers.Logger,
			CloseCh: closeCh,
		})

		require.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	_, body := helpers.HTTPGetBody(t, listenAddr.String(), "/health")
	require.Equal(t, "ok", string(body))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
