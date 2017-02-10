package httpmux

import (
	"net/http"
	"sort"
	"strings"
)

type methodHandler struct {
	all     http.Handler
	methods map[string]http.Handler
}

func newMethodHandler() *methodHandler {
	return &methodHandler{}
}

func (mh *methodHandler) put(handler http.Handler, methods ...string) {
	if len(methods) == 0 {
		mh.all = handler
		return
	}
	if mh.methods == nil {
		mh.methods = map[string]http.Handler{}
	}
	for _, method := range cleanMethods(methods) {
		if handler == nil {
			delete(mh.methods, method)
		} else {
			mh.methods[method] = handler
		}
	}
}

func (mh *methodHandler) get(cleanedMethod string) (http.Handler, error) {
	if mh == nil {
		return nil, ErrNotFound
	}
	if len(mh.methods) == 0 {
		if mh.all != nil {
			return mh.all, nil
		}
		return nil, ErrNotFound
	}
	handler := mh.methods[cleanedMethod]
	if handler != nil {
		return handler, nil
	}
	if mh.all != nil {
		return mh.all, nil
	}
	return nil, ErrMethodNotAllowed(mh.listMethods())
}

func (mh *methodHandler) isRegistered() bool {
	return mh.all != nil || len(mh.methods) > 0
}

func (mh *methodHandler) listMethods() []string {
	result := make([]string, 0, len(mh.methods))
	for method, _ := range mh.methods {
		result = append(result, method)
	}
	sort.Strings(result)
	return result
}

func cleanMethods(methods []string) []string {
	result := make([]string, 0, len(methods))
	for _, method := range methods {
		result = append(result, cleanMethod(method))
	}
	return result
}

func cleanMethod(method string) string {
	return strings.TrimSpace(strings.ToUpper(method))
}
