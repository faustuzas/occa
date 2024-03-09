package test

import (
	"bytes"
	"context"
	"encoding/json"
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

func HTTPGetBody(t testing.TB, addr, path string) (int, []byte) {
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

func HTTPExec(t testing.TB, req *http.Request) (*http.Response, []byte) {
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, body
}

func ToJSONBody(t testing.TB, val any) io.Reader {
	b, err := json.Marshal(val)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func FromJSONBytes[T any](t testing.TB, bytes []byte) T {
	var val T
	require.NoError(t, json.Unmarshal(bytes, &val))
	return val
}
