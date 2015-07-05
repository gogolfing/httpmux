package mux

import (
	"net/http"

	errors "github.com/gogolfing/mux/errors"
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
	return t.root.SubRoute(path).Handle(handler, methods...)
}

func (t *trie) subRoute(path string) *Route {
	return t.root.SubRoute(path)
}

func (t *trie) getHandler(r *http.Request, path string) (http.Handler, error) {
	parent, found, remainingPath := t.root.findSubRoute(path)
	if len(remainingPath) == 0 {
		if found == nil {
			return nil, errors.ErrNotFound
		}
		return found.getHandler(r)
	}
	return parent.getHandler(r)
}
