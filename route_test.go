package httpmux

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	empty    = newRoute("")
	a        = newRoute("a")
	b        = newRoute("b")
	hello    = newRoute("hello")
	another  = newRoute("another")
	boy      = newRoute("boy")
	chooses  = newRoute("chooses")
	division = newRoute("division")
	elephant = newRoute("elephant")
	frogs    = newRoute("frogs")
	giraffe  = newRoute("giraffe")
)

func TestNewRoute(t *testing.T) {
	tests := []struct {
		path     string
		children []*Route
	}{
		{"path", nil},
		{"path", []*Route{}},
		{"another path", []*Route{&Route{}, &Route{}}},
	}
	for _, test := range tests {
		route := newRoute(test.path, test.children...)
		if route.path != test.path {
			t.Errorf("route.path = %v want %v", route.path, test.path)
		}
		if !areRoutesEqual(route.children, test.children) {
			t.Fail()
		}
	}
}

func TestNewRoute_empty(t *testing.T) {
	route := newRoute("/")
	if route.children != nil {
		t.Fail()
	}
}

func TestRoute_methodHandlers(t *testing.T) {
	zero := intHandler(0)
	one := intHandler(1)
	two := intHandler(2)
	three := intHandler(3)
	four := intHandler(4)

	root := newRoute("")
	root.Delete(zero)
	root.Get(one)
	root.Post(two)
	root.Put(three)
	root.Patch(four)

	tests := []struct {
		method  string
		handler http.Handler
	}{
		{"DELETE", zero},
		{"GET", one},
		{"POST", two},
		{"PUT", three},
		{"PATCH", four},
	}
	for _, test := range tests {
		r, _ := http.NewRequest(test.method, "localhost", nil)
		handler, err := root.getHandler(r)
		if handler != test.handler || err != nil {
			t.Errorf("*Route.getHandler(%q) = %v, %v want %v, %v", test.method, handler, err, test.handler, nil)
		}
	}
}

func TestRoute_methodHandlerFuncs(t *testing.T) {
	root := newRoute("")
	root.DeleteFunc(intHandler(0).ServeHTTP)
	root.GetFunc(intHandler(1).ServeHTTP)
	root.PatchFunc(intHandler(4).ServeHTTP)
	root.PostFunc(intHandler(2).ServeHTTP)
	root.PutFunc(intHandler(3).ServeHTTP)

	testRouteResponse(t, root, "DELETE", 200, "0")
	testRouteResponse(t, root, "GET", 200, "1")
	testRouteResponse(t, root, "PATCH", 200, "4")
	testRouteResponse(t, root, "POST", 200, "2")
	testRouteResponse(t, root, "PUT", 200, "3")
}

func TestRoute_HandleFunc(t *testing.T) {
	root := newRoute("")
	handler := intHandler(0)
	result := root.HandleFunc(handler.ServeHTTP)
	if result != root {
		t.Fail()
	}
	testRouteResponse(t, root, "GET", 200, "0")
}

func TestRoute_Handle(t *testing.T) {
	root := newRoute("")
	handler := intHandler(0)
	result := root.Handle(handler)
	if result != root {
		t.Fail()
	}
	r, _ := http.NewRequest("GET", "localhost", nil)
	resultHandler, err := root.getHandler(r)
	if resultHandler != handler || err != nil {
		t.Fail()
	}
}

func testRouteResponse(t *testing.T, route *Route, method string, code int, response string) {
	r, _ := http.NewRequest(method, "localhost", nil)
	handler, err := route.getHandler(r)
	if err != nil {
		t.Fail()
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != code || w.Body.String() != response {
		t.Fail()
	}
}

func TestRoute_getHandler_routeHandleNil(t *testing.T) {
	route := newRoute("")
	handler, err := route.getHandler(nil)
	if handler != nil || err != ErrNotFound {
		t.Fail()
	}
}

func TestRoute_getHandler_useRouteHandler(t *testing.T) {
	handler := intHandler(0)
	route := newRoute("")
	route.routeHandler = &routeHandler{
		handler,
		nil,
	}
	h, err := route.getHandler(nil)
	if h != handler || err != nil {
		t.Fail()
	}
}

func TestRoute_insertStaticSubRoute_empty(t *testing.T) {
	root := newRoute("root")
	result, err := root.insertStaticSubRoute("")
	if result != root || err != nil {
		t.Fail()
	}
}

func TestRoute_insertStaticSubRoute_exists(t *testing.T) {
	root := newRoute("",
		newRoute("hello"),
	)
	result, err := root.insertStaticSubRoute("hello")
	if result != root.children[0] || err != nil {
		t.Fail()
	}
}

func TestRoute_insertStaticSubRoute_leaf(t *testing.T) {
	root := newRoute("",
		newRoute("another"),
		newRoute("hello"),
	)
	result, err := root.insertStaticSubRoute("hello, world")
	if result.path != ", world" || result != root.children[1].children[0] || err != nil {
		t.Fail()
	}
}

func TestRoute_insertStaticSubRoute_splitChild_noRemaining(t *testing.T) {
	root := newRoute("",
		newRoute("hello"),
	)
	result, err := root.insertStaticSubRoute("he")
	if result.path != "he" || len(root.children) != 1 || result != root.children[0] || err != nil {
		t.Fail()
	}
}

func TestRoute_insertStaticSubRoute_splitChild_remaining(t *testing.T) {
	root := newRoute("",
		newRoute("hello"),
	)
	result, err := root.insertStaticSubRoute("hey")
	if result.path != "y" || len(root.children) != 1 || result != root.children[0].children[1] || err != nil {
		t.Error(result, root.children[0])
		t.Fail()
	}
}

func TestRoute_insertChildPath(t *testing.T) {
	tests := []struct {
		root  *Route
		path  string
		index int
	}{
		{newRoute(""), "", -1},
		{newRoute("", newRoute("hello")), "", -1},
		{newRoute(""), "hello", 0},
		{newRoute("", newRoute("hello")), "another", 0},
		{newRoute("", newRoute("hello")), "paper", 1},
	}
	for _, test := range tests {
		result := test.root.insertStaticChildPath(test.path)
		if test.index == -1 {
			if result != test.root {
				t.Error("*Route.insertChildPath(%q) = %v want root", test.path, result)
			}
		} else {
			if result != test.root.children[test.index] {
				t.Fail()
			}
		}
	}
}

func TestRoute_splitStaticChild(t *testing.T) {
	tests := []struct {
		childPaths    []string
		path          string
		index         int
		oldChildPath  string
		newChildPath  string
		remainingPath string
	}{
		{[]string{"hello"}, "he", 0, "llo", "he", ""},
		{[]string{"another", "boy", "chooses"}, "bodacious", 1, "y", "bo", "dacious"},
		{[]string{"another", "bodacious", "chooses"}, "boy", 1, "dacious", "bo", "y"},
	}
	for _, test := range tests {
		children := make([]*Route, 0, len(test.childPaths))
		for _, path := range test.childPaths {
			children = append(children, newRoute(path))
		}
		root := newRoute("", children...)
		oldChild := root.children[test.index]
		result, remainingPath := root.splitStaticChild(test.path)
		if result != root.children[test.index] ||
			remainingPath != test.remainingPath ||
			len(result.children) != 1 ||
			result.children[0] != oldChild ||
			result.children[0].path != test.oldChildPath ||
			result.path != test.newChildPath {
			t.Fail()
		}
	}
}

func TestRoute_findStaticSubRoute(t *testing.T) {
	t.Log("root with no children")
	root := newRoute("")
	testRoute_findStaticSubRoute(t, root, "", root, nil, "")
	testRoute_findStaticSubRoute(t, root, "hello", root, nil, "hello")

	t.Log("root with single child")
	helloRoute := newRoute("hello")
	root = newRoute("", helloRoute)
	testRoute_findStaticSubRoute(t, root, "", root, nil, "")
	testRoute_findStaticSubRoute(t, root, "he", root, helloRoute, "he")
	testRoute_findStaticSubRoute(t, root, "hey", root, helloRoute, "hey")
	testRoute_findStaticSubRoute(t, root, "hello", root, helloRoute, "")
	testRoute_findStaticSubRoute(t, root, "hello, world", helloRoute, nil, ", world")
	testRoute_findStaticSubRoute(t, root, "another", root, nil, "another")

	t.Log("root with single child and single grandchild")
	worldRoute := newRoute(", world")
	helloRoute = newRoute("hello", worldRoute)
	root = newRoute("", helloRoute)
	testRoute_findStaticSubRoute(t, root, "", root, nil, "")
	testRoute_findStaticSubRoute(t, root, "he", root, helloRoute, "he")
	testRoute_findStaticSubRoute(t, root, "hello", root, helloRoute, "")
	testRoute_findStaticSubRoute(t, root, "hey", root, helloRoute, "hey")
	testRoute_findStaticSubRoute(t, root, "hello, world", helloRoute, worldRoute, "")
	testRoute_findStaticSubRoute(t, root, "another", root, nil, "another")
	testRoute_findStaticSubRoute(t, root, "hello, wo", helloRoute, worldRoute, ", wo")
	testRoute_findStaticSubRoute(t, root, "hello, wonderful", helloRoute, worldRoute, ", wonderful")
	testRoute_findStaticSubRoute(t, root, "hello, world, again", worldRoute, nil, ", again")

	t.Log("root with variable sub routes")
	//TODO implement this.
}

func testRoute_findStaticSubRoute(t *testing.T, root *Route, path string, expectedParent, expectedFound *Route, expectedRemainingPath string) {
	parent, found, remainingPath := root.findStaticSubRoute(path)
	if parent != expectedParent || found != expectedFound || remainingPath != expectedRemainingPath {
		t.Errorf("%v.findStaticSubRoute(%q) = %v, %v, %q want %v, %v, %q",
			root, path, parent, found, remainingPath, expectedParent, expectedFound, expectedRemainingPath)
	}
}

func TestRoute_findStaticChildWithCommonPrefix(t *testing.T) {
	tests := []struct {
		path     string
		children []*Route
		child    *Route
		index    int
		prefix   string
	}{
		{"", []*Route{}, nil, -1, ""},
		{"hello", nil, nil, -1, ""},
		{"hello", []*Route{}, nil, -1, ""},
		{"", []*Route{a, b}, nil, -1, ""},
		{"", []*Route{empty, hello}, empty, 0, ""},
		{"character", []*Route{another, boy, chooses, division}, chooses, 2, "ch"},
		{"divisor", []*Route{another, boy, chooses, division}, division, 3, "divis"},
		{"ant", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, another, 0, "an"},
		{"ant", []*Route{boy, chooses, division, elephant, frogs, giraffe}, nil, -1, ""},
		{"hello", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, nil, -8, ""},
		{"boy", []*Route{another, chooses, division, elephant}, nil, -2, ""},
		{"boy", []*Route{another, boy, chooses, division, elephant, frogs}, boy, 1, "boy"},
		{"boys", []*Route{another, boy, chooses}, boy, 1, "boy"},
		//TODO add options with variable child.
	}
	for _, test := range tests {
		route := newRoute("", test.children...)
		child, index, prefix := route.findStaticChildWithCommonPrefix(test.path)
		if child != test.child || index != test.index || prefix != test.prefix {
			t.Errorf("route.findStaticChildWithCommonPrefix(%q) = %v, %v, %q want %v, %v, %q",
				test.path, child, index, prefix, test.child, test.index, test.prefix)
		}
	}
}

func TestRoute_indexOfCommonPrefixChild(t *testing.T) {
	tests := []struct {
		path     string
		children []*Route
		index    int
		prefix   string
	}{
		{"", []*Route{}, -1, ""},
		{"hello", []*Route{}, -1, ""},
		{"", []*Route{a, b}, -1, ""},
		{"", []*Route{empty, hello}, 0, ""},
		{"character", []*Route{another, boy, chooses, division}, 2, "ch"},
		{"divisor", []*Route{another, boy, chooses, division}, 3, "divis"},
		{"ant", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, 0, "an"},
		{"ant", []*Route{boy, chooses, division, elephant, frogs, giraffe}, -1, ""},
		{"hello", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, -8, ""},
		{"boy", []*Route{another, chooses, division, elephant}, -2, ""},
		{"boy", []*Route{another, boy, chooses, division, elephant, frogs}, 1, "boy"},
		{"boys", []*Route{another, boy, chooses}, 1, "boy"},
	}
	for _, test := range tests {
		route := newRoute("", test.children...)
		index, prefix := route.indexOfStaticCommonPrefixChild(test.path)
		if index != test.index || prefix != test.prefix {
			t.Errorf("route.indexOfCommonPrefixChild(%q) = %v, %q want %v, %q",
				test.path, index, prefix, test.index, test.prefix,
			)
		}
	}
}

func TestRoute_insertChildAtIndex(t *testing.T) {
	one := newRoute("one")
	two := newRoute("two")
	three := newRoute("three")
	tests := []struct {
		children       []*Route
		insert         *Route
		index          int
		resultChildren []*Route
		resultReturn   *Route
	}{
		{nil, one, -1, nil, nil},
		{nil, one, 2, nil, nil},
		{[]*Route{}, one, 2, []*Route{}, nil},
		{[]*Route{one}, two, 4, []*Route{one}, nil},
		{nil, one, 0, []*Route{one}, one},
		{[]*Route{}, one, 0, []*Route{one}, one},
		{[]*Route{one}, two, 0, []*Route{two, one}, two},
		{[]*Route{one}, two, 1, []*Route{one, two}, two},
		{[]*Route{one, two}, three, 0, []*Route{three, one, two}, three},
		{[]*Route{one, two}, three, 1, []*Route{one, three, two}, three},
		{[]*Route{one, two}, three, 2, []*Route{one, two, three}, three},
		//these should not occur during normal use, but still testing.
		{[]*Route{one, two}, nil, 1, []*Route{one, nil, two}, nil},
		{[]*Route{one, two}, two, 1, []*Route{one, two, two}, two},
	}
	for _, test := range tests {
		route := newRoute("route", test.children...)
		result := route.insertStaticChildAtIndex(test.insert, test.index)
		if !areRoutesEqual(route.children, test.resultChildren) || result != test.resultReturn {
			t.Errorf("%v insertChildAtIndex(%v, %v) = %v want %v", test.children, test.insert, test.index, route.children, test.resultChildren)
		}
	}
}

func TestRoute_Methods(t *testing.T) {
	root := newRoute("")
	root.Handle(nil, "PUT", "GET")
	methods := root.Methods()
	if !reflect.DeepEqual(methods, []string{"GET", "PUT"}) {
		t.Fail()
	}
}

func TestAreRoutesEqual(t *testing.T) {
	one, two := newRoute("one"), newRoute("two")
	tests := []struct {
		a      []*Route
		b      []*Route
		equals bool
	}{
		{nil, nil, true},
		{nil, []*Route{}, false},
		{[]*Route{}, []*Route{}, true},
		{nil, []*Route{one}, false},
		{[]*Route{one}, []*Route{two}, false},
		{[]*Route{one, two}, []*Route{one, two}, true},
		{[]*Route{one, nil}, []*Route{one, nil}, true},
		{[]*Route{one, nil}, []*Route{one, two}, false},
	}
	for _, test := range tests {
		equals := areRoutesEqual(test.a, test.b)
		if equals != test.equals {
			t.Errorf("areRoutesEqual(%v, %v) = %v want %v", test.a, test.b, equals, test.equals)
		}
	}
}

func areRoutesEqual(a, b []*Route) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
