package path

import "testing"

var PrefixPairs = []struct {
	A string
	B string
}{
	{"", ""},
	{"foo", ""},
	{"", "foo"},
	{"foobar", "foobar"},
	{"abcd", "abef"},
}

func BenchmarkCompareAfterPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, pair := range PrefixPairs {
			CompareAfterPrefix(pair.A, pair.B)
		}
	}
}

func BenchmarkCommonPrefixLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, pair := range PrefixPairs {
			CommonPrefixLen(pair.A, pair.B)
		}
	}
}

func BenchmarkCompareIgnoringPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, pair := range PrefixPairs {
			CompareIgnoringPrefix(pair.A, pair.B)
		}
	}
}
