package mux

import (
	"net/http"
	"sort"
)

type routeHandler struct {
	handler        http.Handler
	methodHandlers map[string]http.Handler
}

func (rh *routeHandler) getHandler(r *http.Request) (http.Handler, error) {
	if len(rh.methodHandlers) == 0 {
		return rh.handler, nil
	}
	if h, ok := rh.methodHandlers[r.Method]; ok {
		return h, nil
	}
	if rh.handler != nil {
		return rh.handler, nil
	}
	return nil, ErrMethodNotAllowed(rh.methods())
}

func (rh *routeHandler) handleFunc(hf http.HandlerFunc, methods ...string) {
	rh.handle(http.HandlerFunc(hf), methods...)
}

func (rh *routeHandler) handle(h http.Handler, methods ...string) {
	if len(methods) == 0 {
		rh.handler = h
		return
	}
	if rh.methodHandlers == nil {
		rh.methodHandlers = make(map[string]http.Handler, len(methods))
	}
	for _, method := range methods {
		rh.methodHandlers[method] = h
	}
}

func (rh *routeHandler) methods() []string {
	result := make([]string, 0, len(rh.methodHandlers))
	for method, _ := range rh.methodHandlers {
		result = append(result, method)
	}
	sort.Strings(result)
	return result
}
