package mux

import (
	"fmt"
	"net/http"
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

func (route *Route) getHandler(r *http.Request) (http.Handler, error) {
	if route.routeHandler == nil {
		return nil, ErrNotFound
	}
	return route.routeHandler.getHandler(r)
}

//func (route *Route) find(path string) (*Route, int, string) {
//	var parent *Route = nil
//	child := route
//	index, prefix := -1, muxpath.CommonPrefix(child.path, path)
//	for len(prefix)
//}
//
//func (route *Route) find(path string) *Route {
//	var found *Route = nil
//	child := route
//	prefix := muxpath.CommonPrefix(child.path, path)
//	for child != nil && len(prefix) == len(child.path) {
//		found = child
//		path = path[len(prefix):]
//		child, _, prefix = found.findChildWithCommonPrefix(path)
//	}
//	//not found OR exact match on path OR no children.
//	if found == nil || len(path) == 0 || len(found.children) == 0 {
//		return found
//	}
//	return nil
//}
//
//func (route *Route) SubRoute(path string) *Route {
//	return route.insertChildPath(path)
//}
//
//func (route *Route) insert(path string) *Route {
//	prefix := muxpath.CommonPrefix(path, route.path)
//	if len(prefix) > 0 {
//		//path shares a prefix with this route.
//		childPath := path[len(prefix):]
//		if len(prefix) == len(route.path) {
//			return route.insertChildPath(childPath)
//		}
//		route.splitPathToPrefix(prefix)
//		return route.insertChildPath(childPath)
//	}
//	//path does not share a prefix with this route.
//	return route.insertChildPath(path)
//}
//
//func (route *Route) insertChildPath(childPath string) *Route {
//	if len(childPath) == 0 {
//		return route
//	}
//	child, _ := route.findOrCreateChildWithCommonPrefix(childPath)
//	return child.insert(childPath)
//}
//
//func (route *Route) splitPathToPrefix(prefix string) {
//	if len(prefix) == 0 {
//		return
//	}
//	childPath := route.path[len(prefix):]
//	if len(childPath) == 0 {
//		return
//	}
//	child := &Route{childPath, route.children, route.routeHandler}
//	route.children = []*Route{child}
//	route.path = prefix
//}
//
//func (route *Route) findOrCreateChildWithCommonPrefix(path string) (*Route, string) {
//	child, index, prefix := route.findChildWithCommonPrefix(path)
//	if child != nil {
//		return child, prefix
//	}
//	child = &Route{path, nil, nil}
//	route.insertChildAtIndex(child, ^index)
//	return child, path
//}
//
//func (route *Route) findChildWithCommonPrefix(path string) (*Route, int, string) {
//	index, prefix := route.indexOfCommonPrefixChild(path)
//	if index >= 0 {
//		return route.children[index], index, prefix
//	}
//	return nil, index, prefix
//}
//
//func (route *Route) indexOfCommonPrefixChild(path string) (int, string) {
//	low, high := 0, len(route.children)
//	for low < high {
//		mid := (low + high) >> 1
//		comparison, prefix := muxpath.CompareIgnorePrefix(path, route.children[mid].path)
//		if len(prefix) > 0 {
//			return mid, prefix
//		} else if comparison == 0 {
//			return mid, path
//		} else if comparison < 0 {
//			high = mid
//		} else { //comparison must be > 0.
//			low = mid + 1
//		}
//	}
//	return ^high, ""
//}
//
//func (route *Route) insertChildAtIndex(child *Route, index int) {
//	if index < 0 || index > len(route.children) {
//		return
//	}
//	if route.children == nil {
//		route.children = []*Route{child}
//		return
//	}
//	before := route.children[:index]
//	after := route.children[index:]
//	route.children = make([]*Route, 0, len(before)+1+len(after))
//	route.children = append(route.children, before...)
//	route.children = append(route.children, child)
//	route.children = append(route.children, after...)
//}
//
func (route *Route) String() string {
	return fmt.Sprintf("&Route{%s}", route.path)
}
