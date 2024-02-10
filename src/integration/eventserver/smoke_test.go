package eventserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/eventserver"
	"github.com/faustuzas/occa/src/integration/serviceboot"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestEventServer_Smoke(t *testing.T) {
	params := serviceboot.DefaultEventServerParams(t)
	go func() {
		require.NoError(t, eventserver.Start(params))
	}()

	require.Eventually(t, func() bool {
		_, body := pkgtest.HTTPGetBody(t, params.HTTPListenAddress.String(), "/health")
		return pkghttp.DefaultOKResponse() == string(body)
	}, time.Second, 100*time.Millisecond)
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
