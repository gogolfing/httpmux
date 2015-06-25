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

func (t *trie) handle(path string, handler http.Handler, methods ...string) *Route {
	return t.root.insertSubRoute(path).Handle(handler, methods...)
}

func (t *trie) subRoute(path string) *Route {
	return t.root.SubRoute(path)
}

func (t *trie) getHandler(r *http.Request, path string) (http.Handler, error) {
	//will likely need paent here for error handling.
	_, found, remainingPath := t.root.findSubRoute(path)
	if len(remainingPath) == 0 {
		return found.getHandler(r)
	}
	return nil, errors.New("something having to do with bad paths")
}
