package httpmux

import (
	"errors"
	"net/http"
)

var (
	errInvalid = errors.New("invalid")
)

type pathType byte

const (
	static pathType = iota
	segmentVar
	endVar
)

func (pt pathType) isStatic() bool {
	return pt == static
}

func (pt pathType) isSegmentVar() bool {
	return pt == segmentVar
}

func (pt pathType) isEndVar() bool {
	return pt == endVar
}

type pathPart struct {
	value string
	pathType
}

type Route struct {
	pathPart

	staticChildren  []*Route
	segmentChildren []*Route
	endChild        *Route

	handler http.Handler
}

func (r *Route) SubRoute(parts ...pathPart) *Route {
	subRoute, err := r.insert(parts)
	if err != nil {
		panic(err)
	}

	return subRoute
}

func (r *Route) insert(parts []pathPart) (*Route, error) {
	if len(parts) == 0 {
		return r, nil
	}

	part := parts[0]
	switch part.pathType {
	case static:
		return r.insertOverlappingStatic(part, parts[1:])

	case segmentVar:
		child := &Route{pathPart: part}
		r.segmentChildren = append(r.segmentChildren, child)
		return child.insert(parts[1:])

	case endVar:
		if len(parts) > 1 {
			return nil, errInvalid
		}
		child := &Route{pathPart: part}
		r.endChild = child
		return r.endChild, nil
	}

	return nil, errInvalid
}

func (r *Route) insertOverlappingStatic(staticPart pathPart, remaining []pathPart) (*Route, error) {
	return nil, nil
}

func commonPrefix(a, b string) string {
	i := 0
	for ; i < len(a) && i < len(b) && a[i] == b[i]; i++ {
	}
	return a[:i]
}

func (r *Route) Handle(h http.Handler) *Route {
	return r
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	panic("unimplemented")
}
