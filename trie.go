package httpmux

import "net/http"

type trie struct {
	root *Route
}

func newTrie() *trie {
	return &trie{
		newRoute(""),
	}
}

func (t *trie) handle(path string, handler http.Handler, methods ...string) *Route {
	return t.subRoute(path).Handle(handler, methods...)
}

func (t *trie) subRoute(path string) *Route {
	return t.root.SubRoute(path)
}

func (t *trie) getHandler(r *http.Request, path string, exact bool) (http.Handler, error) {
	//return nil, errors.ErrNotFound

	//found, _, _ := t.root.searchSubRoute(path, exact)
	//fmt.Println(r.Method, path, vars, remainingPath)
	//return found.getHandler(r)

	handler, _, err := t.root.searchSubRouteHandler(r, path, exact)
	return handler, err
}
