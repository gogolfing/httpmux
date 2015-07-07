package mux

import (
	"net/http"

	errors "github.com/gogolfing/mux/errors"
	muxpath "github.com/gogolfing/mux/path"
)

type Mux struct {
	trie                    *trie
	MethodNotAllowedHandler http.Handler
	NotFoundHandler         http.Handler
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

func (m *Mux) Root() *Route {
	return m.trie.subRoute("/")
}

func (m *Mux) HandleFunc(path string, handlerFunc http.HandlerFunc, methods ...string) *Route {
	return m.Handle(path, http.HandlerFunc(handlerFunc), methods...)
}

func (m *Mux) Handle(path string, handler http.Handler, methods ...string) *Route {
	return m.trie.handle(muxpath.EnsureRootSlash(path), handler, methods...)
}

func (m *Mux) SubRoute(path string) *Route {
	return m.trie.subRoute(muxpath.EnsureRootSlash(path))
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
		return
	}
}

func (m *Mux) serveError(w http.ResponseWriter, r *http.Request, err error) {
	handler := m.getErrorHandler(err)
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (m *Mux) getErrorHandler(err error) http.Handler {
	if handler, ok := err.(errors.ErrMethodNotAllowed); ok {
		if m.MethodNotAllowedHandler != nil {
			return m.MethodNotAllowedHandler
		}
		return handler
	}
	if err == errors.ErrNotFound {
		if m.NotFoundHandler != nil {
			return m.NotFoundHandler
		}
		return errors.ErrNotFound
	}
	return nil
}
