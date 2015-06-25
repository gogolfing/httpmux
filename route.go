package mux

import (
	"net/http"

	muxpath "github.com/gogolfing/mux/path"
)

type Route struct {
	path         string
	children     []*Route
	routeHandler *routeHandler
}

func newRoute(path string, children ...*Route) *Route {
	return &Route{
		path,
		children,
		nil,
	}
}

func (route *Route) Delete(handler http.Handler) *Route {
	return route.Handle(handler)
}

func (route *Route) DeleteFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc)
}

func (route *Route) Get(handler http.Handler) *Route {
	return route.Handle(handler)
}

func (route *Route) GetFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc)
}

func (route *Route) Post(handler http.Handler) *Route {
	return route.Handle(handler)
}

func (route *Route) PostFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc)
}

func (route *Route) Put(handler http.Handler) *Route {
	return route.Handle(handler)
}

func (route *Route) PutFunc(handlerFunc http.HandlerFunc) *Route {
	return route.HandleFunc(handlerFunc)
}

func (route *Route) SubRoute(path string) *Route {
	return route.insertSubRoute(path)
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
		return nil, ErrNotFound
	}
	return route.routeHandler.getHandler(r)
}

func (route *Route) insertSubRoute(path string) *Route {
	parent, found, remainingPath := route.findSubRoute(path)
	if len(remainingPath) == 0 {
		return found
	}
	if found == nil {
		return parent.insertLeaf(remainingPath)
	}
	return parent.insertSplitChild(remainingPath)
}

func (route *Route) insertLeaf(path string) *Route {
	child := newRoute(path)
	route.insertChildAtIndex(child, 0)
	return child
}

func (route *Route) insertSplitChild(path string) *Route {
	oldChild, index, _ := route.findChildWithCommonPrefix(path)
	newChild := newRoute(path, oldChild)
	route.children[index] = newChild
	oldChild.path = oldChild.path[len(path):]
	return newChild
}

func (route *Route) findSubRoute(path string) (*Route, *Route, string) {
	parent := route
	child, _, prefix := parent.findChildWithCommonPrefix(path)
	for child != nil && len(path) > 0 && len(prefix) == len(child.path) {
		path = path[len(prefix):]
		if len(path) == 0 {
			break
		}
		parent = child
		child, _, prefix = parent.findChildWithCommonPrefix(path)
	}
	return parent, child, path
}

func (route *Route) findChildWithCommonPrefix(path string) (*Route, int, string) {
	index, prefix := route.indexOfCommonPrefixChild(path)
	if index >= 0 {
		return route.children[index], index, prefix
	}
	return nil, index, prefix
}

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

func (route *Route) insertChildAtIndex(child *Route, index int) {
	if index < 0 || index > len(route.children) {
		return
	}
	if route.children == nil {
		route.children = []*Route{child}
		return
	}
	before := route.children[:index]
	after := route.children[index:]
	route.children = make([]*Route, 0, len(before)+1+len(after))
	route.children = append(route.children, before...)
	route.children = append(route.children, child)
	route.children = append(route.children, after...)
}

func (route *Route) Methods() []string {
	return route.routeHandler.methods()
}
