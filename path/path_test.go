package path

import (
	"reflect"
	"testing"
)

func TestSplitIntoStaticAndVariableParts(t *testing.T) {
	tests := []struct {
		path   string
		result []string
	}{
		{"", []string{}},
		{" ", []string{" "}},
		{"foobar", []string{"foobar"}},

		{" *", []string{" ", "*"}},
		{"*", []string{"*"}},
		{"*foobar", []string{"*foobar"}},
		{":", []string{":"}},
		{" :", []string{" ", ":"}},
		{":foobar", []string{":foobar"}},

		{"::", []string{":"}},
		{"**", []string{"*"}},
		{"X::", []string{"X:"}},
		{"X**", []string{"X*"}},
		{"::X", []string{":X"}},
		{": :X", []string{": :X"}},
		{"**X", []string{"*X"}},
		{"* *X", []string{"* *X"}},

		{"/:foo/bar", []string{"/", ":foo", "/bar"}},
		{":foo/bar", []string{":foo", "/bar"}},
		{":foo:bar", []string{":foo:bar"}},
		{":foo:bar/else", []string{":foo:bar", "/else"}},
		{"/:foo/:bar/:else/prefix:more", []string{"/", ":foo", "/", ":bar", "/", ":else", "/prefix", ":more"}},
		{"/:::foo", []string{"/:", ":foo"}},
		{"/**:foo", []string{"/*", ":foo"}},
		{`/:_)(*&_%)&#@_) @(&1023495870124:POIHJIOUH______++_+_\\'/`, []string{"/", `:_)(*&_%)&#@_) @(&1023495870124:POIHJIOUH______++_+_\\'`, "/"}},

		{"/*foo/bar", []string{"/", "*foo/bar"}},
		{"*foo/bar", []string{"*foo/bar"}},
		{"*foo:bar", []string{"*foo:bar"}},
		{"*foo*bar/else", []string{"*foo*bar/else"}},
		{"/:foo/:bar/:else/prefix*more", []string{"/", ":foo", "/", ":bar", "/", ":else", "/prefix", "*more"}},
		{"/::*foo", []string{"/:", "*foo"}},
		{"/***foo", []string{"/*", "*foo"}},
		{`/*_)(*&_%)&#@_) @(&1023495870124:POIHJIOUH______++_+_\\'`, []string{"/", `*_)(*&_%)&#@_) @(&1023495870124:POIHJIOUH______++_+_\\'`}},
	}
	for _, test := range tests {
		result := SplitIntoStaticAndVariableParts(test.path)
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("SplitIntoStaticAndVariableParts(%q) = %v WANT %v", test.path, result, test.result)
		}
	}
}

func TestExtractVariableName(t *testing.T) {
	tests := []struct {
		value        string
		resultString string
		resultBool   bool
	}{
		{"", "", false},
		{" ", "", false},
		{"foobar", "", false},
		{" *", "", false},
		{"*", "", true},
		{"*foobar", "foobar", true},
		{"*foo bar", "foo bar", true},
		{":", "", true},
		{":foobar", "foobar", true},
		{":foo bar", "foo bar", true},
	}
	for _, test := range tests {
		resultString, resultBool := ExtractVariableName(test.value)
		if resultString != test.resultString || resultBool != test.resultBool {
			t.Errorf(
				"ExtractVariableName(%q) = %q, %v WANT %q, %v",
				test.value,
				resultString,
				resultBool,
				test.resultString,
				test.resultBool,
			)
		}
	}
}

func TestIsSegmentVariable(t *testing.T) {
	tests := []struct {
		value  string
		result bool
	}{
		{"", false},
		{" ", false},
		{"foobar", false},
		{" *", false},
		{"*", false},
		{"*foobar", false},
		{"*foo bar", false},
		{":", true},
		{":foobar", true},
		{":foo bar", true},
	}
	for _, test := range tests {
		result := IsSegmentVariable(test.value)
		if result != test.result {
			t.Errorf("IsSegmentVariable(%q) = %v WANT %v", test.value, result, test.result)
		}
	}
}

func TestIsEndVariable(t *testing.T) {
	tests :=
		[]struct {
			value  string
			result bool
		}{
			{"", false},
			{" ", false},
			{"foobar", false},
			{" *", false},
			{":", false},
			{":foobar", false},
			{":foo bar", false},
			{"*", true},
			{"*foobar", true},
			{"*foo bar", true},
		}
	for _, test := range tests {
		result := IsEndVariable(test.value)
		if result != test.result {
			t.Errorf("IsEndVariable(%q) = %v WANT %v", test.value, result, test.result)
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
			t.Errorf("Clean(%q) = %q WANT %q", test.path, cleaned, test.cleaned)
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
			t.Errorf("EnsureRootSlash(%q) = %q WANT %q", test.path, result, test.result)
		}
	}
}

/*
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
			t.Errorf("CommonPrefix(%q, %q) = %q WANT %q", test.a, test.b, prefix, test.prefix)
		}
	}
}
*/

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
		comparison, prefix := CompareIgnoringPrefix(test.a, test.b)
		if comparison != test.comparison || prefix != test.prefix {
			t.Errorf("CompareIgnorePrefix(%q, %q) = %v, %q WANT %v, %q",
				test.a, test.b, comparison, prefix, test.comparison, test.prefix)
		}
	}
}
