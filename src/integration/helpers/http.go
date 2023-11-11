package helpers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func HTTPGetBody(t *testing.T, addr, path string) (int, []byte) {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	resp, err := http.Get(addr + path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}
