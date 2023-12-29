package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func RequestLogger(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			var statusCode int
			next.ServeHTTP(capturingResponseWriter{
				ResponseWriter: w,
				statusCode:     &statusCode,
			}, r)

			logger.Info("request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Stringer("duration", time.Since(startTime)),
				zap.Int("status_code", statusCode),
				zap.String("remote_address", r.RemoteAddr))
		})
	}
}

type capturingResponseWriter struct {
	http.ResponseWriter
	statusCode *int
}

// WriteHeader captures the status code before calling the underlying WriteHeader method.
func (w capturingResponseWriter) WriteHeader(code int) {
	*w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
