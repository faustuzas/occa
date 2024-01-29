package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	"github.com/faustuzas/occa/src/integration/containers"
	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGatewaySmoke(t *testing.T) {
	params := serviceboot.DefaultGatewayParams(t, containers.WithMysql(t), containers.WithRedis(t))
	go func() {
		require.NoError(t, gateway.Start(params))
	}()

	require.Eventually(t, func() bool {
		_, body := pkgtest.HTTPGetBody(t, params.ServerListenAddress.String(), "/health")
		return pkghttp.DefaultOKResponse() == string(body)
	}, time.Second, 100*time.Millisecond)
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
