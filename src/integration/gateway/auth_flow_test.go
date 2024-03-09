package gateway

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestGateway_AuthFlow(t *testing.T) {
	params := DefaultParams(t)
	go func() {
		require.NoError(t, gateway.Start(params))
	}()

	tester := NewTester(t, context.Background(), params.HTTPListenAddress.String())

	// Perform authorized call and fail.
	req := tester.createReq(http.MethodPost, "/heartbeat", nil)
	resp, _ := pkgtest.HTTPExec(t, req)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	tester.Register("user", "password")
	token := tester.Login("user", "password")

	// Perform authorized call and succeed.
	req = tester.createReq(http.MethodPost, "/heartbeat", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ = pkgtest.HTTPExec(t, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
