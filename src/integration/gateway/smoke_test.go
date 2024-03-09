package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGateway_Smoke(t *testing.T) {
	params := DefaultParams(t)
	go func() {
		require.NoError(t, gateway.Start(params))
	}()

	require.Eventually(t, func() bool {
		_, body := pkgtest.HTTPGetBody(t, params.HTTPListenAddress.String(), "/health")
		return pkghttp.DefaultOKResponse() == string(body)
	}, time.Second, 100*time.Millisecond)
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
