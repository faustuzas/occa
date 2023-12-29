package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	middlewarestd "github.com/slok/go-http-metrics/middleware/std"
)

func BasicMetrics(r prometheus.Registerer) func(http.Handler) http.Handler {
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Registry: r,
		}),
	})
	
	return func(next http.Handler) http.Handler {
		return middlewarestd.Handler("", mdlw, next)
	}
}
