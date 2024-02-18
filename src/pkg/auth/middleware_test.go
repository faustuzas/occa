package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestHTTPMiddleware_MissingHeader(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		validatorMock = NewMockTokenValidator(ctrl)
	)

	authMiddleware := HTTPTokenAuthorizationMiddleware(pkgtest.Instrumentation, validatorMock)

	passed := false
	srv := httptest.NewServer(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passed = true
	})))
	defer srv.Close()

	status, _ := pkgtest.HTTPGetBody(t, srv.URL, "")
	require.Equal(t, http.StatusUnauthorized, status)
	require.False(t, passed)
}

func TestHTTPMiddleware_Success(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		validatorMock = NewMockTokenValidator(ctrl)

		token     = "a token"
		principal = Principal{ID: pkgid.NewID(), UserName: "mr. test"}
	)

	validatorMock.EXPECT().Validate(gomock.Any(), token).Return(principal, nil)
	authMiddleware := HTTPTokenAuthorizationMiddleware(pkgtest.Instrumentation, validatorMock)

	var actualPrincipal Principal
	srv := httptest.NewServer(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPrincipal = PrincipalFromContext(r.Context())
	})))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := pkgtest.HTTPExec(t, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, principal, actualPrincipal)
}
