package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
)

type principalKey int

var key principalKey

func HTTPMiddleware(l *zap.Logger, validator TokenValidator) httpmiddleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				pkghttp.RespondWithJSONError(l, w, pkghttp.ErrUnauthorized(fmt.Errorf("missing Authorization header")))
				return
			}

			if strings.HasPrefix("Bearer ", token) {
				token = token[len("Bearer "):]
			}

			principal, err := validator.Validate(token)
			if err != nil {
				pkghttp.RespondWithJSONError(l, w, pkghttp.ErrUnauthorized(err))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), key, principal))

			l.Info("received request with Authorization header", zap.String("token", token))
			next.ServeHTTP(w, r)
		})
	}
}

func PrincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(key).(Principal)
}
