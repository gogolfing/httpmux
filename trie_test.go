package mux

import (
	"net/http"
	"testing"
)

func TestNewTrie(t *testing.T) {
	trie := newTrie()
	if trie.root.path != "" {
		t.Fail()
	}
}

func TestTrie_getHandler(t *testing.T) {
	tests := []struct {
		paths   []string
		path    string
		handler http.Handler
		err     error
	}{
		{nil, "", nil, ErrNotFound},
		{[]string{"hello"}, "", nil, ErrNotFound},
		{nil, "hello", nil, ErrNotFound},
		{[]string{"hello"}, "hello", intHandler(0), nil},
		{[]string{"hello"}, "something", nil, ErrNotFound},
		{[]string{"hello"}, "he", nil, ErrNotFound},
		{[]string{"hello"}, "hello, world", intHandler(0), nil},
		{[]string{"romane", "romanus", "romulus", "rubens", "ruber", "rubicon", "rubicundus"}, "ruber", intHandler(4), nil},
		{[]string{"romane", "romanus", "romulus", "rubens", "ruber", "rubicon", "rubicundus"}, "rom", nil, ErrNotFound},
	}
	for _, test := range tests {
		trie := newTrie()
		for i, path := range test.paths {
			trie.handle(path, intHandler(i), "GET")
		}
		r, _ := http.NewRequest("GET", "localhost", nil)
		handler, err := trie.getHandler(r, test.path)
		if handler != test.handler || err != test.err {
			t.Errorf("trie.getHandler(%q) = %v, %v want %v, %v", test.path, handler, err, test.handler, test.err)
		}
	}
}
