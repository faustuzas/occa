package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/eventserver/services"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
)

type Services struct {
	EventServer        services.EventServer
	HTTPAuthMiddleware httpmiddleware.Middleware

	Logger   *zap.Logger
	Registry *prometheus.Registry
}

func Configure(s Services) (http.Handler, error) {
	rawRouter := pkghttp.NewRouterBuilder(s.Logger)
	rawRouter.HandleFunc("/metrics", promhttp.HandlerFor(s.Registry, promhttp.HandlerOpts{}).ServeHTTP).
		Methods(http.MethodGet)

	instrumentedRouter := rawRouter.SubGroup().
		With(httpmiddleware.BasicMetrics(s.Registry), httpmiddleware.RequestLogger(s.Logger))

	instrumentedRouter.HandleJSONFunc("/health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodGet)

	authenticatedRouter := instrumentedRouter.SubGroup().
		With(s.HTTPAuthMiddleware)

	// TODO: protobuf alternative should be added too
	authenticatedRouter.HandleJSONFunc("/send-message", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var msg SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			return nil, err
		}

		if err := s.EventServer.SendEvent(r.Context(), services.Event{
			SenderID:    pkgauth.PrincipalFromContext(r.Context()).ID,
			RecipientID: msg.RecipientID,
			Content:     msg.Content,
			SentAt:      time.Now(),
		}); err != nil {
			return nil, fmt.Errorf("sending message: %w", err)
		}

		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	return rawRouter.Build(), nil
}
