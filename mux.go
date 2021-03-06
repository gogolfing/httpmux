package httpmux

import (
	"context"
	"net/http"

	muxpath "github.com/gogolfing/httpmux/path"
)

type variablesKey int

const variablesKeyValue variablesKey = 1

type Mux struct {
	root *Route

	AllowTrailingSlashes bool

	DisallowSettingAllowMethodHeader bool
	MethodNotAllowedHandler          http.Handler

	NotFoundHandler http.Handler
}

func New() *Mux {
	return &Mux{
		root: newRootRoute(),
		MethodNotAllowedHandler: ErrStatusHandler(http.StatusMethodNotAllowed),
	}
}

func (m *Mux) HandleFunc(path string, handlerFunc http.HandlerFunc, methods ...string) *Route {
	return m.Handle(path, http.HandlerFunc(handlerFunc), methods...)
}

func (m *Mux) Handle(path string, handler http.Handler, methods ...string) *Route {
	return m.SubRoute(path).Handle(handler, methods...)
}

func (m *Mux) Root() *Route {
	return m.root
}

func (m *Mux) SubRoute(path string) *Route {
	return m.root.SubRoute(path)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, vars, err := m.root.findHandler(r, m.getFoundMatcher())
	if err != nil {
		m.serveError(w, r, err)
		return
	}
	r = m.mapVariables(r, vars)
	handler.ServeHTTP(w, r)
}

func (m *Mux) getFoundMatcher() foundMatcher {
	if m.AllowTrailingSlashes {
		return stringFoundMatcher(muxpath.Slash)
	}
	return stringFoundMatcher("")
}

func (m *Mux) serveError(w http.ResponseWriter, r *http.Request, err error) {
	handler := m.getErrorHandler(err)
	if handler == nil {
		return
	}
	handler.ServeHTTP(w, r)
}

func (m *Mux) getErrorHandler(err error) http.Handler {
	if errMNA, ok := err.(ErrMethodNotAllowed); ok {
		var result http.Handler
		if m.MethodNotAllowedHandler != nil {
			result = m.MethodNotAllowedHandler
		}
		if !m.DisallowSettingAllowMethodHeader {
			return &setHeaderHandler{name: HeaderAllow, value: errMNA.Header(), wrapped: result}
		}
		return result
	}
	if err == ErrNotFound {
		if m.NotFoundHandler != nil {
			return m.NotFoundHandler
		}
		return ErrNotFound
	}
	return nil
}

type setHeaderHandler struct {
	name    string
	value   string
	wrapped http.Handler
}

func (shh *setHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(shh.name, shh.value)

	if shh.wrapped != nil {
		shh.wrapped.ServeHTTP(w, r)
	}
}

func (m *Mux) mapVariables(r *http.Request, vars []*Variable) *http.Request {
	if len(vars) == 0 {
		return r
	}

	ctx := context.WithValue(r.Context(), variablesKeyValue, vars)
	for _, v := range vars {
		ctx = context.WithValue(ctx, v.Name, v.Value)
	}
	return r.WithContext(ctx)
}

func VariablesFrom(c context.Context) []*Variable {
	vars, _ := c.Value(variablesKeyValue).([]*Variable)
	return vars
}

func VariableFrom(c context.Context, name string) *Variable {
	v, _ := VariableFromOk(c, name)
	return v
}

func VariableFromOk(c context.Context, name string) (*Variable, bool) {
	value, ok := c.Value(VarName(name)).(string)
	if !ok {
		return nil, ok
	}
	return &Variable{VarName(name), value}, ok
}
