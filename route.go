package httpmux

import (
	"fmt"
	"net/http"

	errors "github.com/gogolfing/httpmux/errors"
	muxpath "github.com/gogolfing/httpmux/path"
)

type pathType uint8

const (
	pathTypeStatic pathType = iota
	pathTypePartVariable
	pathTypeEndVariable
)

func (pt pathType) isVariable() bool {
	return pt.isPartVariable() || pt.IsEndVariable()
}

func (pt pathType) isPartVariable() bool {
	return pt == pathTypePartVariable
}

func (pt pathType) IsEndVariable() bool {
	return pt == pathTypeEndVariable
}

type Route struct {
	path         string
	pathType     pathType
	children     []*Route
	routeHandler *routeHandler
}

func newRoute(path string, pt pathType, children ...*Route) *Route {
	return &Route{
		path,
		pt,
		children,
		nil,
	}
}

func (route *Route) Delete(handler http.Handler) *Route {
	return route.Handle(handler, "DELETE")
}

func (route *Route) DeleteFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc, "DELETE")
}

func (route *Route) Get(handler http.Handler) *Route {
	return route.Handle(handler, "GET")
}

func (route *Route) GetFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc, "GET")
}

func (route *Route) Post(handler http.Handler) *Route {
	return route.Handle(handler, "POST")
}

func (route *Route) PostFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc, "POST")
}

func (route *Route) Put(handler http.Handler) *Route {
	return route.Handle(handler, "PUT")
}

func (route *Route) PutFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc, "PUT")
}

func (route *Route) SubRoute(path string) *Route {
	return route
}

func (route *Route) HandleFunc(handlerFunc http.HandlerFunc, methods ...string) *Route {
	return route.Handle(http.HandlerFunc(handlerFunc), methods...)
}

func (route *Route) Handle(handler http.Handler, methods ...string) *Route {
	if route.routeHandler == nil {
		route.routeHandler = &routeHandler{}
	}
	route.routeHandler.handle(handler, methods...)
	return route
}

func (route *Route) getHandler(r *http.Request) (http.Handler, error) {
	if route.routeHandler == nil {
		return nil, errors.ErrNotFound
	}
	return route.routeHandler.getHandler(r)
}

func (route *Route) insertSubRoute(path string) (*Route, error) {
	if len(path) == 0 {
		return route, nil
	}
	return nil, nil
}

func (route *Route) insertStaticSubRoute(path string) (*Route, error) {
	if len(path) == 0 {
		return route, nil
	}
	if route.pathType.IsEndVariable() {
		return nil, fmt.Errorf(
			"cannot have any path after an end variable: %v %v",
			route.path,
			path,
		)
	}
	if route.hasVariableChild() {
		return nil, fmt.Errorf(
			"cannot have static path: %v at same location as variable: %v",
			path,
			route.variableChildPath(),
		)
	}
	//TODO old way of doing it. add static path as a sub route.
	return nil, nil
}

func (route *Route) insertVariableSubRoute(path string) (*Route, error) {
	if route.pathType.isVariable() {
		return nil, fmt.Errorf(
			"cannot have two immediately consecutive variables: %v, %v",
			route.path,
			path,
		)
	}
	//must have static path type.
	if route.hasVariableChild() {
		if route.variableChildPath() == path {
			return route.children[0], nil
		}
		return nil, fmt.Errorf(
			"cannot have unequal variables at the same location: %v, %v",
			route.variableChildPath(),
			path,
		)
	}
	//must have static children.
	if len(route.children) > 0 {
		return nil, fmt.Errorf(
			"cannot have variable: %v at same location as static path: %v",
			path,
			"..."+route.path+"...",
		)
	}
	//must have empty children.
	child := newRoute(path, pathTypePartVariable)
	return child, nil
}

func (route *Route) hasVariableChild() bool {
	return len(route.children) == 1 && route.children[0].pathType.isVariable()
}

func (route *Route) variableChildPath() string {
	return route.children[0].path
}

//func (route *Route) isVariable() bool {
//	return route.pathType == pathTypePartVariable || route.pathType == pathTypeEndVariable
//}
//
//func (route *Route) isEndVariable() bool {
//	return route.pathType == pathTypeEndVariable
//}

func (route *Route) indexOfCommonPrefixChild(path string) (int, string) {
	low, high := 0, len(route.children)
	for low < high {
		mid := (low + high) >> 1
		comparison, prefix := muxpath.CompareIgnorePrefix(path, route.children[mid].path)
		if len(prefix) > 0 {
			return mid, prefix
		} else if comparison == 0 {
			return mid, path
		} else if comparison < 0 {
			high = mid
		} else { //comparison must be > 0.
			low = mid + 1
		}
	}
	return ^high, ""
}

func (route *Route) insertChildAtIndex(child *Route, index int) *Route {
	if index < 0 || index > len(route.children) {
		return nil
	}
	if route.children == nil {
		route.children = []*Route{child}
		return child
	}
	before := route.children[:index]
	after := route.children[index:]
	route.children = make([]*Route, 0, len(before)+1+len(after))
	route.children = append(route.children, before...)
	route.children = append(route.children, child)
	route.children = append(route.children, after...)
	return child
}

func (route *Route) Methods() []string {
	return route.routeHandler.methods()
}
