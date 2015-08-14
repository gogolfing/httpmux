package httpmux

import (
	errorslib "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	errors "github.com/gogolfing/httpmux/errors"
	muxpath "github.com/gogolfing/httpmux/path"
)

const (
	ResponseMethodNotAllowed = "Method Not Allowed\n"
	ResponseNotFound         = "Not Found\n"
)

var (
	ErrUnknown = errorslib.New("unknown error type")
)

func TestNew(t *testing.T) {
	m := New()
	if m.NotFoundHandler != nil || m.MethodNotAllowedHandler != nil {
		t.Fail()
	}
}

func TestNewHandlers(t *testing.T) {
	methodNotAllowed := intHandler(0)
	notFound := intHandler(1)
	m := NewHandlers(methodNotAllowed, notFound)
	if m.MethodNotAllowedHandler != methodNotAllowed || m.NotFoundHandler != notFound {
		t.Fail()
	}
}

func TestMux_Handle(t *testing.T) {
	m := New()
	m.Handle("something", intHandler(0))
	m.Handle("/something/else", intHandler(1))
	m.Handle("/getonly", intHandler(2), "GET")
	m.Handle("/patch", intHandler(3), "PATCH")
	m.Handle("/nil", nil)
	tests := []serveHTTPTest{
		{"GET", "/", http.StatusNotFound, ResponseNotFound},
		{"GET", "/something", http.StatusOK, "0"},
		{"POST", "/something", http.StatusOK, "0"},
		{"GET", "/something/more", http.StatusOK, "0"},
		{"GET", "/something/else", http.StatusOK, "1"},
		{"GET", "/notfound", http.StatusNotFound, ResponseNotFound},
		{"POST", "/getonly", http.StatusMethodNotAllowed, ResponseMethodNotAllowed},
		{"PATCH", "/patch", http.StatusOK, "3"},
		{"GET", "/nil", http.StatusNotFound, ResponseNotFound},
	}
	testMux_ServeHTTP(t, m, tests)
}

func TestMux_HandleFunc(t *testing.T) {
	m := New()
	m.HandleFunc("something", intHandler(0).ServeHTTP)
	m.HandleFunc("/something/else", intHandler(1).ServeHTTP)
	m.HandleFunc("/getonly", intHandler(2).ServeHTTP, "GET")
	m.HandleFunc("/patch", intHandler(3).ServeHTTP, "PATCH")
	tests := []serveHTTPTest{
		{"GET", "/", http.StatusNotFound, ResponseNotFound},
		{"GET", "/something", http.StatusOK, "0"},
		{"POST", "/something", http.StatusOK, "0"},
		{"GET", "/something/more", http.StatusOK, "0"},
		{"GET", "/something/else", http.StatusOK, "1"},
		{"GET", "/notfound", http.StatusNotFound, ResponseNotFound},
		{"POST", "/getonly", http.StatusMethodNotAllowed, ResponseMethodNotAllowed},
		{"PATCH", "/patch", http.StatusOK, "3"},
	}
	testMux_ServeHTTP(t, m, tests)
}

func TestMux_SubRoute(t *testing.T) {
	m := New()
	m.SubRoute("sub").
		Delete(intHandler(0)).
		Get(intHandler(1)).
		Handle(intHandler(2))
	m.Handle("/something", intHandler(3))
	m.SubRoute("/sub/again").SubRoute("/finally").Post(intHandler(4))
	tests := []serveHTTPTest{
		{"GET", "/", http.StatusNotFound, ResponseNotFound},
		{"DELETE", "/sub", http.StatusOK, "0"},
		{"GET", "/sub", http.StatusOK, "1"},
		{"POST", "/sub", http.StatusOK, "2"},
		{"GET", "/sub/else", http.StatusOK, "1"},
		{"POST", "/sub/else", http.StatusOK, "2"},
		{"GET", "/something", http.StatusOK, "3"},
		{"GET", "/sub/again", http.StatusNotFound, ResponseNotFound},
		{"POST", "/sub/again/finally", http.StatusOK, "4"},
		{"GET", "/sub/again/finally", http.StatusMethodNotAllowed, ResponseMethodNotAllowed},
		{"POST", "/sub/again/finally/more", http.StatusOK, "4"},
		{"GET", "/sub/again/finally/more", http.StatusMethodNotAllowed, ResponseMethodNotAllowed},
		{"GET", "/sub/again/else", http.StatusNotFound, ResponseNotFound},
	}
	testMux_ServeHTTP(t, m, tests)
}

func TestMux_SubRoute_variable(t *testing.T) {
	m := New()
	m.SubRoute("/users/{user_id}").Get(intHandler(1))
	tests := []serveHTTPTest{
		{"GET", "/users/1", http.StatusOK, "1"},
		{"GET", "/users/something/else", http.StatusOK, "1"},
	}
	testMux_ServeHTTP(t, m, tests)
}

func TestMux_Handle_root(t *testing.T) {
}

func TestMux_Handle_overwrite(t *testing.T) {
}

type serveHTTPTest struct {
	method   string
	path     string
	code     int
	response string
}

func (s *serveHTTPTest) URL() string {
	return fmt.Sprintf("http://localhost%v", muxpath.EnsureRootSlash(s.path))
}

func testMux_ServeHTTP(t *testing.T, m *Mux, tests []serveHTTPTest) {
	for _, test := range tests {
		r, _ := http.NewRequest(test.method, test.URL(), nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, r)
		response := w.Body.String()
		if w.Code != test.code || response != test.response {
			t.Errorf("%v.ServeHTTP(%q, %q) = %v, %q want %v, %q", m, test.method, test.path, w.Code, response, test.code, test.response)
		}
	}
}

func TestMux_serveError(t *testing.T) {
	m := New()
	tests := []struct {
		err      error
		code     int
		response string
	}{
		{errors.ErrMethodNotAllowed([]string{"GET"}), http.StatusMethodNotAllowed, ResponseMethodNotAllowed},
		{errors.ErrNotFound, http.StatusNotFound, ResponseNotFound},
		{ErrUnknown, 200, ""}, //not semantically correct but is result of empty response.
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
	err := ErrUnknown
	handler := m.getErrorHandler(err)
	if handler != nil {
		t.Fail()
	}
}
