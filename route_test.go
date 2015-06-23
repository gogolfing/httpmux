package mux

import (
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

func TestRoute_insertSubRoute_exists(t *testing.T) {
	root := newRoute("",
		newRoute("hello"),
	)
	result := root.SubRoute("hello")
	if result != root.children[0] {
		t.Fail()
	}
}

func TestRoute_insertSubRoute_leaf(t *testing.T) {
	root := newRoute("",
		newRoute("another"),
		newRoute("hello"),
	)
	result := root.SubRoute("hello, world")
	if result.path != ", world" || result != root.children[1].children[0] {
		t.Fail()
	}
}

func TestRoute_insertSubRoute_splitChild(t *testing.T) {
	root := newRoute("",
		newRoute("hello"),
	)
	result := root.SubRoute("he")
	if result.path != "he" || len(root.children) != 1 || result != root.children[0] {
		t.Fail()
	}
}

func TestRoute_getHandler(t *testing.T) {
}

func TestRoute_insertLeaf(t *testing.T) {
	const helloString = "hello"
	root := newRoute("")
	result := root.insertLeaf(helloString)
	if result.path != helloString || len(root.children) != 1 || root.children[0] != result {
		t.Errorf("%v.insertLeaf(%q) = %q, %v want %q, %v",
			root, helloString, result.path, root.children, helloString, []*Route{newRoute(helloString)})
	}
}

func TestRoute_insertSplitChild(t *testing.T) {
	tests := []struct {
		root    *Route
		oldPath string
		newPath string
	}{
		{
			newRoute("",
				newRoute("hello"),
			),
			"hello",
			"he",
		},
		{
			newRoute("",
				newRoute("another"),
				newRoute("baseball"),
				newRoute("car"),
				newRoute("diamond"),
				newRoute("hello"),
				newRoute("world"),
			),
			"hello",
			"hell",
		},
	}
	for _, test := range tests {
		expectedNumberChildren := len(test.root.children)
		_, expectedOldRoute, remainingPath := test.root.findSubRoute(test.oldPath)
		_, expectedIndex, _ := test.root.findChildWithCommonPrefix(test.oldPath)
		if expectedOldRoute == nil || len(remainingPath) > 0 {
			t.Errorf("route.findSubRoute(%q) = %v, %q is incorrect", test.oldPath, expectedOldRoute, remainingPath)
		}
		expectedOldRoute.routeHandler = &routeHandler{}
		expectedNewRoute := test.root.insertSplitChild(test.newPath)

		_, oldRoute, remaingPath := test.root.findSubRoute(test.oldPath)
		if oldRoute != expectedOldRoute || len(remaingPath) > 0 || !reflect.DeepEqual(oldRoute, expectedOldRoute) {
			t.Errorf("route.insertSplitChild(%q) oldRoute = %v want %v", test.oldPath, oldRoute, expectedOldRoute)
		}
		_, newRoute, remaingPath := test.root.findSubRoute(test.newPath)
		if newRoute != expectedNewRoute || newRoute.path != test.newPath || len(remainingPath) > 0 {
			t.Errorf("route.insertSplitChild(%q) newRoute = %v want %v", test.newPath, newRoute, expectedNewRoute)
		}
		_, index, _ := test.root.findChildWithCommonPrefix(test.newPath)
		if index != expectedIndex {
			t.Errorf("route.insertSplitChild(%q) index = %v want %v", test.newPath, index, expectedIndex)
		}
		numberChildren := len(test.root.children)
		if expectedNumberChildren != numberChildren {
			t.Errorf("route.insertSplitChild(%q) number of children = %v want %v", test.newPath, numberChildren, expectedNumberChildren)
		}
	}
}

func TestRoute_findSubRoute(t *testing.T) {
	t.Log("root with no children")
	root := newRoute("")
	testRoute_findSubRoute(t, root, "", root, nil, "")
	testRoute_findSubRoute(t, root, "hello", root, nil, "hello")

	t.Log("root with single child")
	helloRoute := newRoute("hello")
	root = newRoute("", helloRoute)
	testRoute_findSubRoute(t, root, "", root, nil, "")
	testRoute_findSubRoute(t, root, "he", root, helloRoute, "he")
	testRoute_findSubRoute(t, root, "hello", root, helloRoute, "")
	testRoute_findSubRoute(t, root, "hello, world", helloRoute, nil, ", world")
	testRoute_findSubRoute(t, root, "another", root, nil, "another")

	t.Log("root with single child and single grandchild")
	worldRoute := newRoute(", world")
	helloRoute = newRoute("hello", worldRoute)
	root = newRoute("", helloRoute)
	testRoute_findSubRoute(t, root, "", root, nil, "")
	testRoute_findSubRoute(t, root, "he", root, helloRoute, "he")
	testRoute_findSubRoute(t, root, "hello", root, helloRoute, "")
	testRoute_findSubRoute(t, root, "hello, world", helloRoute, worldRoute, "")
	testRoute_findSubRoute(t, root, "another", root, nil, "another")
	testRoute_findSubRoute(t, root, "hello, wo", helloRoute, worldRoute, ", wo")
	testRoute_findSubRoute(t, root, "hello, world, again", worldRoute, nil, ", again")
}

func testRoute_findSubRoute(t *testing.T, root *Route, path string, expectedParent, expectedFound *Route, expectedRemainingPath string) {
	parent, found, remainingPath := root.findSubRoute(path)
	if parent != expectedParent || found != expectedFound || remainingPath != expectedRemainingPath {
		t.Errorf("%v.findSubRoute(%q) = %v, %v, %q want %v, %v, %q",
			root, path, parent, found, remainingPath, expectedParent, expectedFound, expectedRemainingPath)
	}
}

//func TestRoute_listAllPaths(t *testing.T) {
//	root := &Route{
//		"a",
//		[]*Route{
//			newRoute("b"),
//			&Route{
//				"c",
//				[]*Route{newRoute("d"), newRoute("e")},
//				nil,
//			},
//		},
//		nil,
//	}
//	tests := []struct {
//		root   *Route
//		result []string
//	}{
//		{root, []string{"a", "ab", "ac", "acd", "ace"}},
//		{newRoute("nil children"), []string{"nil children"}},
//		{&Route{"simple", []*Route{newRoute("a"), newRoute("b")}, nil}, []string{"simple", "simplea", "simpleb"}},
//		{&Route{"a", []*Route{&Route{"b", []*Route{newRoute("c")}, nil}}, nil}, []string{"a", "ab", "abc"}},
//	}
//	for _, test := range tests {
//		result := test.root.listAllPaths()
//		if !reflect.DeepEqual(result, test.result) {
//			t.Errorf("%v.listAllPaths() = %v want %v", test.root, result, test.result)
//		}
//	}
//}
//
//func (route *Route) listAllPaths() []string {
//	result := []string{route.path}
//	for _, child := range route.children {
//		childPaths := child.listAllPaths()
//		for _, childPath := range childPaths {
//			result = append(result, route.path+childPath)
//		}
//	}
//	return result
//}
//
//func TestLevelOrder(t *testing.T) {
//	leafNil := newRoute("leafNil")
//	leafEmpty := &Route{"leafEmpty", []*Route{}, nil}
//	singleChild := &Route{"singleChild", []*Route{division}, nil}
//	multiChild := &Route{"multiChild", []*Route{another, boy, chooses}, nil}
//	bigsChild := &Route{"bigsChild", []*Route{leafEmpty, singleChild}, nil}
//	big := &Route{"big", []*Route{leafNil, bigsChild, multiChild, frogs}, nil}
//	tests := []struct {
//		root   *Route
//		result []*Route
//	}{
//		{leafNil, []*Route{leafNil}},
//		{leafEmpty, []*Route{leafEmpty}},
//		{singleChild, []*Route{singleChild, division}},
//		{bigsChild, []*Route{bigsChild, leafEmpty, singleChild, division}},
//		{multiChild, []*Route{multiChild, another, boy, chooses}},
//		{big, []*Route{big, leafNil, bigsChild, multiChild, frogs, leafEmpty, singleChild, another, boy, chooses, division}},
//	}
//	for _, test := range tests {
//		result := test.root.levelOrder()
//		if !areRoutesEqual(result, test.result) {
//			t.Errorf("%v.levelOrder() = %v want %v", test.root, result, test.result)
//		}
//	}
//}
//
//func (route *Route) levelOrder() []*Route {
//	result := []*Route{}
//	queue := []*Route{route}
//	for len(queue) > 0 {
//		temp := queue[0]
//		result = append(result, temp)
//		queue = append(queue, temp.children...)
//		queue = queue[1:]
//	}
//	return result
//}
//
func TestRoute_findChildWithCommonPrefix(t *testing.T) {
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
	}
	for _, test := range tests {
		route := newRoute("", test.children...)
		child, index, prefix := route.findChildWithCommonPrefix(test.path)
		if child != test.child || index != test.index || prefix != test.prefix {
			t.Errorf("route.findChildWithCommonPrefix(%q) = %v, %v, %q want %v, %v, %q",
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
		index, prefix := route.indexOfCommonPrefixChild(test.path)
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
		children []*Route
		insert   *Route
		index    int
		result   []*Route
	}{
		{nil, one, -1, nil},
		{nil, one, 2, nil},
		{[]*Route{}, one, 2, []*Route{}},
		{[]*Route{one}, two, 4, []*Route{one}},
		{nil, one, 0, []*Route{one}},
		{[]*Route{}, one, 0, []*Route{one}},
		{[]*Route{one}, two, 0, []*Route{two, one}},
		{[]*Route{one}, two, 1, []*Route{one, two}},
		{[]*Route{one, two}, three, 0, []*Route{three, one, two}},
		{[]*Route{one, two}, three, 1, []*Route{one, three, two}},
		{[]*Route{one, two}, three, 2, []*Route{one, two, three}},
		//these should not occur during normal use, but still testing.
		{[]*Route{one, two}, nil, 1, []*Route{one, nil, two}},
		{[]*Route{one, two}, two, 1, []*Route{one, two, two}},
	}
	for _, test := range tests {
		route := &Route{"route", test.children, nil}
		route.insertChildAtIndex(test.insert, test.index)
		equals := areRoutesEqual(route.children, test.result)
		if !equals {
			t.Errorf("%v insertChildAtIndex(%v, %v) = %v want %v", test.children, test.insert, test.index, route.children, test.result)
		}
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
