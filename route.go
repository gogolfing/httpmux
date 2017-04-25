package httpmux

import (
	"net/http"

	muxpath "github.com/gogolfing/httpmux/path"
)

type Route struct {
	node
}

func newRootRoute() *Route {
	return newRoute(
		&staticNode{},
	)
}

func newRoute(n node) *Route {
	return &Route{
		node: n,
	}
}

func (r *Route) DeleteFunc(handlerFunc http.HandlerFunc) *Route {
	return r.Delete(handlerFunc)
}

func (r *Route) Delete(handler http.Handler) *Route {
	return r.Handle(handler, http.MethodDelete)
}

func (r *Route) GetFunc(handlerFunc http.HandlerFunc) *Route {
	return r.Get(handlerFunc)
}

func (r *Route) Get(handler http.Handler) *Route {
	return r.Handle(handler, http.MethodGet)
}

func (r *Route) PatchFunc(handlerFunc http.HandlerFunc) *Route {
	return r.Patch(handlerFunc)
}

func (r *Route) Patch(handler http.Handler) *Route {
	return r.Handle(handler, http.MethodPatch)
}

func (r *Route) PostFunc(handlerFunc http.HandlerFunc) *Route {
	return r.Post(handlerFunc)
}

func (r *Route) Post(handler http.Handler) *Route {
	return r.Handle(handler, http.MethodPost)
}

func (r *Route) PutFunc(handlerFunc http.HandlerFunc) *Route {
	return r.Put(handlerFunc)
}

func (r *Route) Put(handler http.Handler) *Route {
	return r.Handle(handler, http.MethodPut)
}

func (r *Route) Handle(handler http.Handler, methods ...string) *Route {
	r.node.put(handler, methods...)
	return r
}

func (r *Route) SubRoute(path string) *Route {
	resultNode := r.node
	var err error = nil

	path = muxpath.Clean(path)
	parts := muxpath.SplitIntoStaticAndVariableParts(path)
	for _, part := range parts {
		name, ok := muxpath.ExtractVariableName(part)
		if ok {
			switch {
			case muxpath.IsSegmentVariable(part):
				resultNode, err = resultNode.appendSegmentVar(VarName(name))
			case muxpath.IsEndVariable(part):
				resultNode, err = resultNode.appendEndVar(VarName(name))
			}
		} else {
			resultNode, err = resultNode.appendStatic(part)
		}

		if err != nil {
			panic(err)
		}
	}

	if resultNode == r.node {
		return r
	}

	return newRoute(resultNode)
}

func (r *Route) findHandler(req *http.Request, m foundMatcher) (http.Handler, []*Variable, error) {
	found, vars := r.node.find(muxpath.Clean(req.URL.Path), m)

	if found == nil {
		return nil, nil, ErrNotFound
	}

	handler, err := found.get(req.Method)
	if err != nil {
		return nil, nil, err
	}
	return handler, vars, nil
}
