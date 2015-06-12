package path

import "testing"

func TestClean(t *testing.T) {
	tests := []struct {
		path    string
		cleaned string
	}{
		{"", "/"},
		{"", "/"},
		{"/", "/"},
		{"/.", "/"},
		{"/../../", "/"},
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

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		path   string
		result bool
	}{
		{"", true},
		{"hello, world", false},
		{"golfmux", false},
	}
	for _, test := range tests {
		isEmpty := IsEmpty(test.path)
		if isEmpty != test.result {
			t.Errorf("IsEmpty(%q) = %v want %v", test.path, isEmpty, test.result)
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
	}
	for _, test := range tests {
		comparison, prefix := CompareIgnorePrefix(test.a, test.b)
		if comparison != test.comparison || prefix != test.prefix {
			t.Errorf("CompareIgnorePrefix(%q, %q) = %v, %q want %v, %q",
				test.a, test.b, comparison, prefix, test.comparison, test.prefix)
		}
	}
}
