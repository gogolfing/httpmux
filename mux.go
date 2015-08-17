package httpmux

import (
	"net/http"
	"sync"

	errors "github.com/gogolfing/httpmux/errors"
	muxpath "github.com/gogolfing/httpmux/path"
)

type Mapper interface {
	Set(*http.Request, []*Variable)
	Delete(*http.Request)
}

type Mux struct {
	trie *Route

	*sync.RWMutex
	varMap map[*http.Request][]*Variable
	Mapper Mapper

	MethodNotAllowedHandler http.Handler
	NotFoundHandler         http.Handler
}

func New() *Mux {
	return NewHandlers(nil, nil)
}

func NewHandlers(notFound, methodNotAllowed http.Handler) *Mux {
	return &Mux{
		trie:                    newRoute(""),
		RWMutex:                 &sync.RWMutex{},
		varMap:                  map[*http.Request][]*Variable{},
		Mapper:                  nil,
		MethodNotAllowedHandler: methodNotAllowed,
		NotFoundHandler:         notFound,
	}
}

func (m *Mux) Root() *Route {
	return m.SubRoute("/")
}

func (m *Mux) HandleFunc(path string, handlerFunc http.HandlerFunc, methods ...string) *Route {
	return m.Handle(path, http.HandlerFunc(handlerFunc), methods...)
}

func (m *Mux) Handle(path string, handler http.Handler, methods ...string) *Route {
	return m.SubRoute(muxpath.EnsureRootSlash(path)).Handle(handler, methods...)
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
	if handler != nil {
		m.mapVariables(r, vars)
		defer m.unmapVariables(r)
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

func (m *Mux) mapVariables(r *http.Request, vars []*Variable) {
	m.Lock()
	defer m.Unlock()
	m.varMap[r] = vars
	if m.Mapper != nil {
		m.Mapper.Set(r, vars)
	}
}

func (m *Mux) unmapVariables(r *http.Request) {
	m.Lock()
	defer m.Unlock()
	delete(m.varMap, r)
	if m.Mapper != nil {
		m.Mapper.Delete(r)
	}
}

func (m *Mux) Vars(r *http.Request) []*Variable {
	m.RLock()
	defer m.RUnlock()
	return m.varMap[r]
}
