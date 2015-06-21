package mux

import "net/http"

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
	return nil, &ErrMethodNotAllowed{
		rh.methods(),
	}
}

func (rh *routeHandler) methods() []string {
	return []string{}
}
