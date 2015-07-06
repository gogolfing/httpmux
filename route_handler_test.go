package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	errors "github.com/gogolfing/mux/errors"
)

type intHandler int

func (h intHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprint(int(h))))
}

func TestRouteHandler_getHandler(t *testing.T) {
	zero := intHandler(0)
	tests := []struct {
		handler        http.Handler
		methods        []string
		methodsHandler http.Handler
		method         string
		result         http.Handler
		err            error
	}{
		{nil, nil, nil, "GET", nil, errors.ErrNotFound},
		{zero, nil, zero, "GET", zero, nil},
		{zero, nil, nil, "GET", nil, errors.ErrNotFound}, //this overwrites rh.handler.
		{nil, []string{"GET"}, nil, "GET", nil, nil},
		{nil, []string{"GET"}, zero, "GET", zero, nil},
		{zero, []string{"GET"}, nil, "PUT", zero, nil},
		{nil, []string{"GET", "POST"}, zero, "PUT", nil, errors.ErrMethodNotAllowed([]string{"GET", "POST"})},
	}
	for _, test := range tests {
		rh := &routeHandler{
			test.handler,
			nil,
		}
		rh.handle(test.methodsHandler, test.methods...)
		r, _ := http.NewRequest(test.method, "localhost", nil)
		result, err := rh.getHandler(r)
		if result != test.result {
			t.Errorf("%v.getHandler(%q) = %v, %v want %v, %v", rh, r.Method, result, err, test.result, test.err)
		}
		if err != nil {
			errMethods, ok := err.(errors.ErrMethodNotAllowed)
			if ok {
				actual := []string(errMethods)
				expected := []string(test.err.(errors.ErrMethodNotAllowed))
				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("%v.getHandler(%q) ErrMethodNotAllowed methods = %v want %v", rh, r.Method, actual, expected)
				}
			} else {
				if err != test.err {
					t.Errorf("%v.getHandler(%q) err = %v want %v", rh, r.Method, err, test.err)
				}
			}
		}
	}
}

func TestRouteHandler_handleFunc(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handleFunc"))
	}
	rh := &routeHandler{}
	rh.handleFunc(f, "GET")
	if rh.handler != nil ||
		len(rh.methodHandlers) != 1 ||
		rh.methodHandlers["GET"].(http.HandlerFunc) == nil {
		t.Fail()
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "localhost", nil)
	rh.methodHandlers["GET"].ServeHTTP(w, r)
	if w.Body.String() != "handleFunc" {
		t.Fail()
	}
}

func TestRouteHandler_handle_noMethods(t *testing.T) {
	rh := &routeHandler{}
	handler := intHandler(0)
	rh.handle(handler)
	if rh.handler != handler || rh.methodHandlers != nil {
		t.Fail()
	}
}

func TestRouteHandler_handle_noMethodOverwrite(t *testing.T) {
	rh := &routeHandler{
		intHandler(0),
		nil,
	}
	rh.handle(nil)
	if rh.handler != nil {
		t.Fail()
	}
}

func TestRouteHandler_handle_methods(t *testing.T) {
	rh := routeHandler{}
	handler := intHandler(0)
	rh.handle(handler, "GET", "PUT")
	if rh.handler != nil ||
		len(rh.methodHandlers) != 2 ||
		rh.methodHandlers["GET"] != handler ||
		rh.methodHandlers["PUT"] != handler {
		t.Fail()
	}
}

func TestRouteHandler_methods(t *testing.T) {
	tests := []struct {
		methods []string
		result  []string
	}{
		{nil, []string{}},
		{[]string{}, []string{}},
		{[]string{"GET"}, []string{"GET"}},
		{[]string{"PUT", "POST", "GET", "DELETE"}, []string{"DELETE", "GET", "POST", "PUT"}},
	}
	for _, test := range tests {
		rh := &routeHandler{}
		rh.handle(nil, test.methods...)
		result := rh.methods()
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("%v.methods() = %v want %v", rh, result, test.result)
		}
	}
}
