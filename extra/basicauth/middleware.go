package basicauth

import (
	"net/http"
	"strconv"

	"github.com/golangwebLK/webrouter"
)

type middleware struct {
	realm string
	check func(req bunlk.Request) (bool, error)
}

type Option func(m *middleware)

func WithRealm(realm string) Option {
	return func(m *middleware) {
		m.realm = strconv.Quote(realm)
	}
}

func NewMiddleware(
	check func(req bunlk.Request) (bool, error),
	opts ...Option,
) bunlk.MiddlewareFunc {
	c := &middleware{
		realm: "Restricted",
		check: check,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (m *middleware) Middleware(next bunlk.HandlerFunc) bunlk.HandlerFunc {
	return func(w http.ResponseWriter, req bunlk.Request) error {
		ok, err := m.check(req)
		if err != nil {
			return err
		}
		if ok {
			return next(w, req)
		}

		return m.basicAuth(w, req)
	}
}

func (m *middleware) basicAuth(w http.ResponseWriter, req bunlk.Request) error {
	w.Header().Set("WWW-Authenticate", "basic realm="+m.realm)
	w.WriteHeader(http.StatusUnauthorized)
	return nil
}
