package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	pkgerrors "github.com/faustuzas/occa/src/pkg/errors"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
)

type principalKey int

var key principalKey

func HTTPTokenAuthorizationMiddleware(l *zap.Logger, validator TokenValidator) httpmiddleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				pkghttp.RespondWithJSONError(l, w, pkgerrors.ErrUnauthorized(fmt.Errorf("missing Authorization header")))
				return
			}

			if strings.HasPrefix(token, "Bearer ") {
				token = token[len("Bearer "):]
			}

			principal, err := validator.Validate(r.Context(), token)
			if err != nil {
				pkghttp.RespondWithJSONError(l, w, pkgerrors.ErrUnauthorized(err))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), key, principal))
			next.ServeHTTP(w, r)
		})
	}
}

func PrincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(key).(Principal)
}

func NoopMiddleware() httpmiddleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), key, noopPrincipal))
			next.ServeHTTP(w, r)
		})
	}
}
