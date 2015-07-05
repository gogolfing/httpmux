package mux

import (
	errorslib "errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	errors "github.com/gogolfing/mux/errors"
)

func TestMux_serveError(t *testing.T) {
	m := New()
	tests := []struct {
		err      error
		code     int
		response string
	}{
		{errors.ErrNotFound, http.StatusNotFound, http.StatusText(http.StatusNotFound) + "\n"},
		{errorslib.New("unknown error type"), 200, ""}, //not semantically correct but is result of empty response.
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "localhost", nil)
		m.serveError(w, r, test.err)
		response := w.Body.String()
		if w.Code != test.code || response != test.response {
			t.Errorf("*Mux.serveError(_, _, %v) = %v, %q want %v, %q", test.err, w.Code, response, test.code, test.response)
		}
	}
}

func TestMux_getErrorHandler_methodNotAllowed(t *testing.T) {
	m := New()
	err := errors.ErrMethodNotAllowed([]string{"GET"})

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
	err := errors.ErrNotFound

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
	err := errorslib.New("unknown error type")
	handler := m.getErrorHandler(err)
	if handler != nil {
		t.Fail()
	}
}
