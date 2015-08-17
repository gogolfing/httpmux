package httpmux

import (
	"fmt"
	"net/http"

	errors "github.com/gogolfing/httpmux/errors"
	muxpath "github.com/gogolfing/httpmux/path"
)

type Variable struct {
	Name  string
	Value string
}

func (v *Variable) String() string {
	return fmt.Sprintf("&%v", *v)
}

type Route struct {
	path     string
	pathType muxpath.PathType

	children []*Route

	varChild *Route

	routeHandler *routeHandler
}

func newRoute(path string, children ...*Route) *Route {
	return &Route{
		path:         path,
		pathType:     muxpath.TypeOf(path),
		children:     children,
		routeHandler: nil,
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
	result, err := route.insertSubRoute(path)
	if err != nil {
		panic(err)
	}
	return result
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

func (route *Route) ListRoutes() []string {
	result := []string{}
	methodsRoutes := route.listMethodsRoutes()
	for _, v := range methodsRoutes {
		result = append(result, v[0]+" "+v[1])
	}
	return result
}

func (route *Route) listMethodsRoutes() [][]string {
	result := [][]string{}
	if route.routeHandler != nil {
		methodsAll := route.routeHandler.methodsAll()
		for _, v := range methodsAll {
			result = append(result, []string{v, route.path})
		}
	}
	for _, child := range route.children {
		methodsRoutes := child.listMethodsRoutes()
		for _, v := range methodsRoutes {
			result = append(result, []string{v[0], route.path + v[1]})
		}
	}
	return result
}

func (route *Route) search(path string) (*Route, []*Variable, string) {
	vars := []*Variable{}
	parent := route
	child, tempVar, remainingPath := parent.searchChildren(path)
	for child != nil {
		if tempVar != nil {
			vars = append(vars, tempVar)
		}
		path = remainingPath
		parent = child
		child, tempVar, remainingPath = parent.searchChildren(path)
	}
	return parent, vars, remainingPath
}

func (route *Route) searchChildren(path string) (*Route, *Variable, string) {
	if route.hasPartVariableChild() {
		return route.searchPartVariableChild(path)
	}
	if route.hasEndVariableChild() {
		return route.searchPartVariableChild(path)
	}
	return route.searchStaticChildren(path)
}

func (route *Route) searchPartVariableChild(path string) (*Route, *Variable, string) {
	name, value, remainingPath := muxpath.ParseVariable(route.variableChildPath(), path)
	return route.varChild, &Variable{name, value}, remainingPath
}

//func (route *Route) searchEndVariableChild(path string) {
//	if route.hasStaticChildren() {
//
//	}
//}

func (route *Route) searchStaticChildren(path string) (*Route, *Variable, string) {
	found, _, prefix := route.findStaticChildWithCommonPrefix(path)
	if found != nil && len(found.path) == len(prefix) {
		return found, nil, path[len(prefix):]
	}
	return nil, nil, path
}

func (route *Route) insertSubRoute(path string) (*Route, error) {
	if len(path) == 0 {
		return route, nil
	}
	result := route
	var err error = nil
	parts := muxpath.SplitPathVars(path)
	isVariable := muxpath.IsVariable(parts[0])
	for _, part := range parts {
		if isVariable {
			result, err = result.insertVariableChild(part)
		} else {
			result, err = result.insertStaticSubRoute(part)
		}
		if err != nil {
			return nil, err
		}
		isVariable = !isVariable
	}
	return result, nil
}

func (route *Route) insertVariableChild(variable string) (*Route, error) {
	if route.isVariable() {
		return nil, &errors.ErrConsecutiveVars{route.path, variable}
	}
	//must have static path type.
	if route.hasVariableChild() {
		if route.variableChildPath() == variable {
			return route.varChild, nil
		}
		return nil, &errors.ErrUnequalVars{route.variableChildPath(), variable}
	}
	//must not have a variable child.
	if route.hasStaticChildren() && muxpath.IsPartVariable(variable) {
		return nil, &errors.ErrOverlapStaticVar{variable, "..." + route.path + "..."}
	}
	//must have empty children OR variable to insert is an end variable.
	route.varChild = newRoute(variable)
	return route.varChild, nil
}

func (route *Route) insertStaticSubRoute(path string) (*Route, error) {
	if len(path) == 0 {
		return route, nil
	}
	if route.isEndVariable() {
		return nil, &errors.ErrOverlapStaticVar{route.path, path}
	}
	if route.hasPartVariableChild() {
		return nil, &errors.ErrOverlapStaticVar{path, route.variableChildPath()}
	}
	parent, found, remainingPath := route.findStaticSubRoute(path)
	if len(remainingPath) == 0 {
		return found, nil
	}
	if parent.hasPartVariableChild() {
		return nil, &errors.ErrOverlapStaticVar{remainingPath, parent.variableChildPath()}
	}
	if found == nil {
		return parent.insertStaticChildPath(remainingPath), nil
	}
	found, remainingPath = parent.splitStaticChild(remainingPath)
	return found.insertStaticChildPath(remainingPath), nil
}

func (route *Route) insertStaticChildPath(path string) *Route {
	if len(path) == 0 {
		return route
	}
	index, _ := route.indexOfStaticCommonPrefixChild(path)
	return route.insertStaticChildAtIndex(newRoute(path), ^index)
}

func (route *Route) splitStaticChild(path string) (*Route, string) {
	oldChild, index, prefix := route.findStaticChildWithCommonPrefix(path)
	newChild := newRoute(prefix, oldChild)
	route.children[index] = newChild
	oldChild.path = oldChild.path[len(prefix):]
	return newChild, path[len(prefix):]
}

func (route *Route) findStaticSubRoute(path string) (*Route, *Route, string) {
	parent := route
	child, _, prefix := parent.findStaticChildWithCommonPrefix(path)
	for child != nil && len(path) > 0 && len(prefix) == len(child.path) {
		path = path[len(prefix):]
		if len(path) == 0 {
			break
		}
		parent = child
		child, _, prefix = parent.findStaticChildWithCommonPrefix(path)
	}
	return parent, child, path
}

func (route *Route) findStaticChildWithCommonPrefix(path string) (*Route, int, string) {
	if route.hasVariableChild() {
		return nil, -1, path
	}
	index, prefix := route.indexOfStaticCommonPrefixChild(path)
	if index >= 0 {
		return route.children[index], index, prefix
	}
	return nil, index, prefix
}

func (route *Route) hasPartVariableChild() bool {
	return route.hasVariableChild() && route.varChild.isPartVariable()
}

func (route *Route) hasEndVariableChild() bool {
	return route.hasVariableChild() && route.varChild.isEndVariable()
}

func (route *Route) hasVariableChild() bool {
	return route.varChild != nil
}

func (route *Route) hasStaticChildren() bool {
	return len(route.children) > 0
}

func (route *Route) variableChildPath() string {
	return route.varChild.path
}

func (route *Route) isVariable() bool {
	return route.pathType.IsVariable()
}

func (route *Route) isPartVariable() bool {
	return route.pathType.IsPartVariable()
}

func (route *Route) isEndVariable() bool {
	return route.pathType.IsEndVariable()
}

func (route *Route) indexOfStaticCommonPrefixChild(path string) (int, string) {
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

func (route *Route) insertStaticChildAtIndex(child *Route, index int) *Route {
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
	if route.routeHandler != nil {
		return route.routeHandler.methods()
	}
	return []string{}
}
