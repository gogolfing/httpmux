package mux

import (
	"net/http"

	muxpath "github.com/gogolfing/mux/path"
)

type Mux struct {
	trie                    *trie
	notFoundHandler         http.Handler
	methodNotAllowedHandler http.Handler
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

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := muxpath.Clean(r.URL.Path)
	handler, err := m.getHandler(r, path)
	if err != nil {
		m.serveError(w, r, err)
		return
	}
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (m *Mux) getHandler(r *http.Request, path string) (http.Handler, error) {
	return m.trie.getHandler(r, path)
}

func (m *Mux) serveError(w http.ResponseWriter, r *http.Request, err error) {
	handler := m.getHandlerFromError(err)
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (m *Mux) getHandlerFromError(err error) http.Handler {
	if methodError, ok := err.(*ErrMethodNotAllowed); ok {
		if m.methodNotAllowedHandler != nil {
			return m.methodNotAllowedHandler
		}
		return methodError
	}
	if err == ErrNotFound {
		if m.notFoundHandler != nil {
			return m.notFoundHandler
		}
		return ErrStatusHandler(http.StatusNotFound)
	}
	return nil
}

func (m *Mux) SetNotFoundHandler(handler http.Handler) {
	m.notFoundHandler = handler
}

func (m *Mux) SetMethodNotAllowedHandler(handler http.Handler) {
	m.methodNotAllowedHandler = handler
}
