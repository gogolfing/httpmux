package mux

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMux(t *testing.T) {
	m := New()
	m.Handle("/", intHandler(0))
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	m.ServeHTTP(w, r)
	t.Log(w.Code, w.Body.String())
	if w.Body.String() != "0" {
		t.Fail()
	}
}

func TestMux_getErrorHandler_methodNotAllowed(t *testing.T) {
	m := New()
	err := ErrMethodNotAllowed([]string{"GET"})

	handler := m.getErrorHandler(err)
	if reflect.DeepEqual(handler, []string{"GET"}) {
		t.Fail()
	}

	m = New()
	m.MethodNotAllowedHandler = intHandler(0)
	handler = m.getErrorHandler(err)
	if handler != m.MethodNotAllowedHandler {
		t.Fail()
	}
}

func TestMux_getErrorHandler_notFound(t *testing.T) {
	m := New()
	err := ErrNotFound

	handler := m.getErrorHandler(err)
	if handler != err {
		t.Fail()
	}

	m = New()
	m.NotFoundHandler = intHandler(0)
	handler = m.getErrorHandler(err)
	if handler != m.NotFoundHandler {
		t.Fail()
	}
}

func TestMux_getErrorHandler_nil(t *testing.T) {
	m := New()
	err := errors.New("unknown error type")
	handler := m.getErrorHandler(err)
	if handler != nil {
		t.Fail()
	}
}
