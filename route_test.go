package golfmux

import "testing"

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
	route := newRoute("path")
	if route.path != "path" {
		t.Errorf("route.path = %v want %v", route.path, "path")
	}
	if route.children != nil {
		t.Errorf("route.children = %v want %v", route.children, nil)
	}
	if route.routeHandler != nil {
		t.Errorf("route.routeHandler = %v want %v", route.routeHandler, nil)
	}
}

func TestRoute_FindOrCreateChildWithCommonPrefix(t *testing.T) {
	tests := []struct {
		path     string
		children []*Route
		child    *Route
		prefix   string
		created  bool
	}{
		{"", []*Route{}, nil, "", true},
		{"hello", []*Route{}, nil, "hello", true},
		{"", []*Route{a, b}, nil, "", true},
		{"", []*Route{empty, hello}, empty, "", false},
		{"character", []*Route{another, boy, chooses, division}, chooses, "ch", false},
		{"divisor", []*Route{another, boy, chooses, division}, division, "divis", false},
		{"ant", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, another, "an", false},
		{"ant", []*Route{boy, chooses, division, elephant, frogs, giraffe}, nil, "ant", true},
		{"hello", []*Route{another, boy, chooses, division, elephant, frogs, giraffe}, nil, "hello", true},
		{"boy", []*Route{another, chooses, division, elephant}, nil, "boy", true},
		{"boy", []*Route{another, boy, chooses, division, elephant, frogs}, boy, "boy", false},
		{"boys", []*Route{another, boy, chooses}, boy, "boy", false},
	}
	for _, test := range tests {
		route := &Route{"route", test.children, nil}
		child, prefix := route.findOrCreateChildWithCommonPrefix(test.path)
		passed := prefix == test.prefix && child != nil && (test.created || child == test.child)
		if !passed {
			t.Errorf("route.findOrCreateChildWithCommonPrefix(%q) = %v, %q want %v, %q (%v)",
				test.path, child, prefix, test.child, test.prefix, test.created)
		}
		if test.created {
			index, _ := route.indexOfCommonPrefixChild(test.path)
			if route.children[index] != child {
				t.Errorf("route.findOrCreateChildWithCommonPrefix() did not appropriately create child")
			}
		}
	}
}

func TestRoute_FindChildWithCommonPrefix(t *testing.T) {
	tests := []struct {
		path     string
		children []*Route
		child    *Route
		index    int
		prefix   string
	}{
		{"", []*Route{}, nil, -1, ""},
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
		route := &Route{"route", test.children, nil}
		child, index, prefix := route.findChildWithCommonPrefix(test.path)
		if child != test.child || index != test.index || prefix != test.prefix {
			t.Errorf("route.findChildWithCOmmonPrefix(%q) = %v, %v, %q want %v, %v, %q",
				test.path, child, index, prefix, test.child, test.index, test.prefix)
		}
	}
}

func TestRoute_IndexOfCommonPrefixChild(t *testing.T) {
	tests := []struct {
		path       string
		childPaths []string
		index      int
		prefix     string
	}{
		{"", []string{}, -1, ""},
		{"hello", []string{}, -1, ""},
		{"", []string{"a", "b"}, -1, ""},
		{"", []string{"", "hello"}, 0, ""},
		{"character", []string{"another", "boy", "chooses", "division"}, 2, "ch"},
		{"divisor", []string{"another", "boy", "chooses", "division"}, 3, "divis"},
		{"ant", []string{"another", "boy", "chooses", "division", "elephant", "frogs", "giraffe"}, 0, "an"},
		{"ant", []string{"boy", "chooses", "division", "elephant", "frogs", "giraffe"}, -1, ""},
		{"hello", []string{"another", "boy", "chooses", "division", "elephant", "frogs", "giraffe"}, -8, ""},
		{"boy", []string{"another", "chooses", "division", "elephant"}, -2, ""},
		{"boy", []string{"another", "boy", "chooses", "division", "elephant", "frogs"}, 1, "boy"},
		{"boys", []string{"another", "boy", "chooses"}, 1, "boy"},
	}
	for _, test := range tests {
		route := &Route{"", makeChildrenWithPaths(test.childPaths), nil}
		index, prefix := route.indexOfCommonPrefixChild(test.path)
		if index != test.index || prefix != test.prefix {
			t.Errorf("route.indexOfCommonPrefixChild(%q) = %v, %q want %v, %q",
				test.path, index, prefix, test.index, test.prefix,
			)
		}
	}
}

func TestRoute_InsertChildAtIndex(t *testing.T) {
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

func makeChildrenWithPaths(childPaths []string) []*Route {
	result := make([]*Route, 0)
	for _, path := range childPaths {
		result = append(result, &Route{path, nil, nil})
	}
	return result
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
