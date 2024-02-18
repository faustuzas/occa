package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/gateway/services"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	esmembership "github.com/faustuzas/occa/src/pkg/eventserver/membership"
	"github.com/faustuzas/occa/src/pkg/generated/proto/rteventspb"
	pkghttp "github.com/faustuzas/occa/src/pkg/http"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
)

type Services struct {
	UsersRegisterer     pkgauth.Registerer
	AuthMiddleware      httpmiddleware.Middleware
	ActiveUsersTracker  services.ActiveUsersTracker
	EventServerSelector esmembership.ServerSelector
	RTEventRelay        services.RealTimeEventRelay

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

	instrumentedRouter.HandleJSONFunc("/register", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		err := s.UsersRegisterer.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			return nil, err
		}

		return RegistrationResponse{}, nil
	}).Methods(http.MethodPost)

	instrumentedRouter.HandleJSONFunc("/login", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		token, err := s.UsersRegisterer.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			return nil, err
		}

		return LoginResponse{
			Token: token,
		}, nil
	}).Methods(http.MethodPost)

	authenticatedRouter := instrumentedRouter.SubGroup().
		With(s.AuthMiddleware)

	authenticatedRouter.HandleJSONFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if err := s.ActiveUsersTracker.HeartBeat(r.Context()); err != nil {
			return nil, fmt.Errorf("hearth beating user: %w", err)
		}
		return pkghttp.DefaultOKResponse(), nil
	}).Methods(http.MethodPost)

	authenticatedRouter.HandleJSONFunc("/active-users", func(w http.ResponseWriter, r *http.Request) (any, error) {
		users, err := s.ActiveUsersTracker.ActiveUsers(r.Context())
		if err != nil {
			return nil, fmt.Errorf("getting active users: %w", err)
		}

		return ActiveUsersResponse{
			ActiveUsers: users,
		}, nil
	}).Methods(http.MethodGet)

	authenticatedRouter.HandleJSONFunc("/select-server", func(w http.ResponseWriter, r *http.Request) (any, error) {
		server, err := s.EventServerSelector.SelectServerForConnection(r.Context())
		if err != nil {
			return nil, fmt.Errorf("selecting event server: %w", err)
		}

		principal := pkgauth.PrincipalFromContext(r.Context())

		s.Logger.Info("selected server for user",
			zap.Stringer("userId", principal.ID),
			zap.String("username", principal.UserName),
			zap.String("serverId", server.ID))

		return SelectServerResponse{Address: server.GRPCAddress}, nil
	}).Methods(http.MethodGet)

	authenticatedRouter.HandleJSONFunc("/send-message", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}

		var (
			principal = pkgauth.PrincipalFromContext(r.Context())
			message   = rteventspb.NewDirectMessageEvent(principal.ID, req.Message)
		)
		if err := s.RTEventRelay.Forward(r.Context(), req.RecipientID, message); err != nil {
			return nil, fmt.Errorf("forwaring real time event: %w", err)
		}

		return pkghttp.DefaultOKResponse(), nil
	})

	return rawRouter.Build(), nil
}
