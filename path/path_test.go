package path

import (
	"reflect"
	"testing"
)

func TestSplitPathVars(t *testing.T) {
	tests := []struct {
		path   string
		result []string
	}{
		{"", []string{""}},
		{"something", []string{"something"}},
		{"{hello}", []string{"{hello}"}},
		{"some{var}thing", []string{"some", "{var}", "thing"}},
		{"{var}thing", []string{"{var}", "thing"}},
		{"some{var}", []string{"some", "{var}"}},
		{"{some}{thing}", []string{"{some}", "", "{thing}"}},
		{"so{a}me{b}th{c}ing", []string{"so", "{a}", "me", "{b}", "th", "{c}", "ing"}},
		{"some{a}{b}thing", []string{"some", "{a}", "", "{b}", "thing"}},
		{"some{1234}thing", []string{"some{1234}thing"}},
		{"some{*thing}", []string{"some", "{*thing}"}},
	}
	for _, test := range tests {
		result := SplitPathVars(test.path)
		if len(result) == 0 && len(test.result) == 0 {
			continue
		}
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("SplitPathVars(%q) = %v want %v", test.path, result, test.result)
		}
	}
}

func TestIsEndVariable(t *testing.T) {
	tests := []struct {
		path   string
		result bool
	}{
		//true cases.
		{"{*var}", true},
		{"{*THISISAVARIABLE}", true},
		//false cases.
		{"{var}", false},
		{"{THISISAVARIABLE}", false},
		{"{ var }", false},
		{"something", false},
		{"{1234}", false},
		{"some{thing}", false},
		{"{some}thing", false},
	}
	for _, test := range tests {
		result := IsEndVariable(test.path)
		if result != test.result {
			t.Errorf("IsEndVariable(%q) = %v want %v", test.path, result, test.result)
		}
	}
}

func TestIsVariable(t *testing.T) {
	tests := []struct {
		path   string
		result bool
	}{
		//true cases.
		{"{var}", true},
		{"{THISISAVARIABLE}", true},
		//false cases.
		{"{ var }", false},
		{"something", false},
		{"{1234}", false},
		{"some{thing}", false},
		{"{some}thing", false},
	}
	for _, test := range tests {
		result := IsVariable(test.path)
		if result != test.result {
			t.Errorf("IsVariable(%q) = %v want %v", test.path, result, test.result)
		}
	}
}

func TestFindAllVarSubmatchIndex(t *testing.T) {
	tests := []struct {
		path   string
		result [][]int
	}{
		//will find variables.
		{"something", [][]int{}},
		{"some{var}thing", [][]int{[]int{4, 9}}},
		{"some{el{se}thing", [][]int{[]int{7, 11}}},
		{"some{el}se}thing", [][]int{[]int{4, 8}}},
		{"some{e{l}se}thing", [][]int{[]int{6, 9}}},
		{"so{a}me{b}th{C}ing{D}", [][]int{[]int{2, 5}, []int{7, 10}, []int{12, 15}, []int{18, 21}}},
		{"some{*thing}", [][]int{[]int{4, 12}}},
		//will not find variables.
		{"some{}thing", [][]int{}},
		{"some{var:}thing", [][]int{}},
		{"some{1234}thing", [][]int{}},
		{"{}something", [][]int{}},
		{"something{}", [][]int{}},
		{"so{}me{}th{}ing", [][]int{}},
	}
	for _, test := range tests {
		result := FindAllVarSubmatchIndex(test.path)
		if len(result) == 0 && len(test.result) == 0 {
			continue
		}
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("Regexp Function(%q) = %v want %v", test.path, result, test.result)
		}
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		path    string
		cleaned string
	}{
		{"", "/"},
		{"/", "/"},
		{"/.", "/"},
		{"/../../", "/"},
		{".", "/"},
		{"..", "/"},
		{"./", "/"},
		{"../", "/"},
		{"hello", "/hello"},
		{"/hello", "/hello"},
		{"/hello/", "/hello/"},
		{"hello/", "/hello/"},
		{"hello/./world", "/hello/world"},
		{"hello/../world", "/world"},
		{"hello/..", "/"},
		{"hello/world/.", "/hello/world"},
		{"hello/world/./", "/hello/world/"},
		{"hello/world/..", "/hello"},
		{"hello/world/../", "/hello/"},
	}
	for _, test := range tests {
		cleaned := Clean(test.path)
		if cleaned != test.cleaned {
			t.Errorf("Clean(%q) = %q want %q", test.path, cleaned, test.cleaned)
		}
	}
}

func TestEnsureRootSlash(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"", "/"},
		{"/", "/"},
		{"/hello", "/hello"},
		{"hello/", "/hello/"},
	}
	for _, test := range tests {
		result := EnsureRootSlash(test.path)
		if result != test.result {
			t.Errorf("EnsureRootSlash(%q) = %q want %q", test.path, result, test.result)
		}
	}
}

func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		a      string
		b      string
		prefix string
	}{
		{"", "", ""},
		{"", "hello", ""},
		{"hello", "", ""},
		{"hello, world", "hello", "hello"},
		{"hello", "hello, world", "hello"},
		{"house", "home", "ho"},
		{"hello, world", "golfmux", ""},
	}
	for _, test := range tests {
		prefix := CommonPrefix(test.a, test.b)
		if prefix != test.prefix {
			t.Errorf("CommonPrefix(%q, %q) = %q want %q", test.a, test.b, prefix, test.prefix)
		}
	}
}

func TestCompareIgnorePrefix(t *testing.T) {
	tests := []struct {
		a          string
		b          string
		comparison int
		prefix     string
	}{
		{"", "", 0, ""},
		{"", "hello", -5, ""},
		{"hello", "", 5, ""},
		{"hello", "hello", 0, "hello"},
		{"abc", "abd", -1, "ab"},
		{"abd", "abc", 1, "ab"},
		{"hello", "golfmux", 1, ""},
		{"a", "e", -4, ""},
		{"he", "hello", 0, "he"},
		{"hello", "he", 0, "he"},
	}
	for _, test := range tests {
		comparison, prefix := CompareIgnorePrefix(test.a, test.b)
		if comparison != test.comparison || prefix != test.prefix {
			t.Errorf("CompareIgnorePrefix(%q, %q) = %v, %q want %v, %q",
				test.a, test.b, comparison, prefix, test.comparison, test.prefix)
		}
	}
}
