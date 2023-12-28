package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var _ http.Handler = (*Router)(nil)

type Router struct {
	mux mux.Router

	logger *zap.Logger
}

func NewRouter(l *zap.Logger) *Router {
	return &Router{
		logger: l,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) Use(middlewares ...func(src http.Handler) http.Handler) {
	mws := make([]mux.MiddlewareFunc, len(middlewares))
	for i := range mws {
		mws[i] = middlewares[i]
	}

	r.mux.Use(mws...)
}

func (r *Router) HandleJSONFunc(path string, f func(http.ResponseWriter, *http.Request) (any, error)) *mux.Route {
	return r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		respObj, err := f(w, req)
		if err != nil {
			RespondWithJSONError(r.logger, w, err)
			return
		}
		RespondWithJSON(r.logger, w, respObj)
	})
}
