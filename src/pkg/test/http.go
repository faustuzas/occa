package test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testTimeout = 1 * time.Second
)

func HTTPGetBody(t require.TestingT, addr, path string) (int, []byte) {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), testTimeout)
	defer cancelFn()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr+path, nil)
	require.NoError(t, err)

	resp, body := HTTPExec(t, req)

	return resp.StatusCode, body
}

func HTTPExec(t require.TestingT, req *http.Request) (*http.Response, []byte) {
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, body
}
