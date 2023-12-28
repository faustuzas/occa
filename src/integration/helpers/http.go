package helpers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testTimeout = 1 * time.Second
)

func HTTPGetBody(t *testing.T, addr, path string) (int, []byte) {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), testTimeout)
	defer cancelFn()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}
