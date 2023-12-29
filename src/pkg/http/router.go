package http

import (
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type RouterBuilder struct {
	mux           *mux.Router
	handlersAdded bool

	middlewares []func(next http.Handler) http.Handler

	logger *zap.Logger
}

func NewRouterBuilder(l *zap.Logger) *RouterBuilder {
	gorilla := mux.NewRouter()
	gorilla.Methods(http.MethodGet).Subrouter()

	return &RouterBuilder{
		mux:    mux.NewRouter(),
		logger: l,
	}
}

// SubGroup is used to scope middlewares to some specific routes.
func (b *RouterBuilder) SubGroup() *RouterBuilder {
	return &RouterBuilder{
		mux: b.mux,

		middlewares: slices.Clone(b.middlewares),
		logger:      b.logger,
	}
}

func (b *RouterBuilder) HandleJSONFunc(path string, f func(http.ResponseWriter, *http.Request) (any, error)) *mux.Route {
	b.handlersAdded = true
	return b.mux.Handle(path, b.wrap(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		respObj, err := f(w, req)
		if err != nil {
			RespondWithJSONError(b.logger, w, err)
			return
		}
		RespondWithJSON(b.logger, w, respObj)
	})))
}

func (b *RouterBuilder) HandleFunc(path string, h http.HandlerFunc) *mux.Route {
	return b.mux.Handle(path, b.wrap(h))
}

func (b *RouterBuilder) wrap(h http.Handler) http.Handler {
	for _, m := range b.middlewares {
		h = m(h)
	}
	return h
}

func (b *RouterBuilder) With(middlewares ...func(src http.Handler) http.Handler) *RouterBuilder {
	if b.handlersAdded {
		panic("cannot add middlewares once some handlers are configured")
	}

	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

func (b *RouterBuilder) Build() http.Handler {
	return b.mux
}
