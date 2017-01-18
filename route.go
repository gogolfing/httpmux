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

func (r *Route) findHandler(path, cleanedMethod string, m foundMatcher) (http.Handler, []*Variable, error) {
	path = muxpath.Clean(path)
	found, vars, remaining := r.node.find(path, m)

	if !m.matches(found, remaining) {
		return nil, nil, ErrNotFound
	}

	handler, err := found.get(cleanedMethod)
	if err != nil {
		return nil, nil, err
	}
	return handler, vars, nil
}
