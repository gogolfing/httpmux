package mux

import (
	"errors"
	"net/http"
)

type trie struct {
	root *Route
}

func newTrie() *trie {
	return &trie{
		newRoute(""),
	}
}

func (t *trie) getHandler(r *http.Request, path string) (http.Handler, error) {
	_, found, remainingPath := t.find(path)
	if len(remainingPath) == 0 {
		return found.routeHandler.getHandler(r)
	}
	return nil, errors.New("something having to do with bad paths")
}

func (t *trie) insert(path string) *Route {
	return nil
}

func (t *trie) find(path string) (*Route, *Route, string) {
	return nil, nil, ""
}
