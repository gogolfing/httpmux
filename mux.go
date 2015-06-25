package mux

import (
	"net/http"

	muxpath "github.com/gogolfing/mux/path"
)

type Mux struct {
	trie                    *trie
	NotFoundHandler         http.Handler
	MethodNotAllowedHandler http.Handler
}

func New() *Mux {
	return NewHandlers(nil, nil)
}

func NewHandlers(notFound, methodNotAllowed http.Handler) *Mux {
	return &Mux{
		newTrie(),
		notFound,
		methodNotAllowed,
	}
}

func (m *Mux) HandleFunc(path string, handlerFunc http.HandlerFunc, methods ...string) *Route {
	return m.Handle(path, http.HandlerFunc(handlerFunc), methods...)
}

func (m *Mux) Handle(path string, handler http.Handler, methods ...string) *Route {
	return m.trie.handle(m.ensureRootSlash(path), handler, methods...)
}

func (m *Mux) SubRoute(path string) *Route {
	return m.trie.subRoute(m.ensureRootSlash(path))
}

func (m *Mux) ensureRootSlash(path string) string {
	if len(path) == 0 {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := muxpath.Clean(r.URL.Path)
	handler, err := m.trie.getHandler(r, path)
	if err != nil {
		m.serveError(w, r, err)
		return
	}
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (m *Mux) serveError(w http.ResponseWriter, r *http.Request, err error) {
	handler := m.getErrorHandler(err)
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (m *Mux) getErrorHandler(err error) http.Handler {
	if methodError, ok := err.(*ErrMethodNotAllowed); ok {
		if m.MethodNotAllowedHandler != nil {
			return m.MethodNotAllowedHandler
		}
		return methodError
	}
	if err == ErrNotFound {
		if m.NotFoundHandler != nil {
			return m.NotFoundHandler
		}
		return ErrStatusHandler(http.StatusNotFound)
	}
	return nil
}
