package httpmux

import (
	"context"
	"net/http"

	muxpath "github.com/gogolfing/httpmux/path"
)

type Mux struct {
	trie Route

	MethodNotAllowedHandler http.Handler
	NotFoundHandler         http.Handler
}

func (m *Mux) HandleFunc(path string, handlerFunc http.HandlerFunc, methods ...string) *Route {
	return m.Handle(path, http.HandlerFunc(handlerFunc), methods...)
}

func (m *Mux) Handle(path string, handler http.Handler, methods ...string) *Route {
	return m.SubRoute(muxpath.EnsureRootSlash(path)).Handle(handler, methods...)
}

func (m *Mux) Root() *Route {
	return m.SubRoute(muxpath.RootPath)
}

func (m *Mux) SubRoute(path string) *Route {
	return m.trie.SubRoute(muxpath.EnsureRootSlash(path))
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := muxpath.Clean(r.URL.Path)
	handler, vars, err := m.trie.searchSubRouteHandler(r, path, true)
	if err != nil {
		m.serveError(w, r, err)
		return
	}
	if handler == nil {
		return
	}
	r = m.mapVariables(r, vars)
	handler.ServeHTTP(w, r)
}

func (m *Mux) serveError(w http.ResponseWriter, r *http.Request, err error) {
	handler := m.getErrorHandler(err)
	if handler == nil {
		return
	}
	handler.ServeHTTP(w, r)
}

func (m *Mux) getErrorHandler(err error) http.Handler {
	if handler, ok := err.(ErrMethodNotAllowed); ok {
		if m.MethodNotAllowedHandler != nil {
			return m.MethodNotAllowedHandler
		}
		return handler
	}
	if err == ErrNotFound {
		if m.NotFoundHandler != nil {
			return m.NotFoundHandler
		}
		return ErrNotFound
	}
	return nil
}

func (m *Mux) mapVariables(r *http.Request, vars []*Variable) *http.Request {
	if len(vars) == 0 {
		return r
	}

	ctx := r.Context()
	for _, v := range vars {
		ctx = context.WithValue(ctx, v.Name, v.Value)
	}
	return r.WithContext(ctx)
}
