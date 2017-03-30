package httpmux

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	NotFoundBody = "Not Found\n"
)

type TestResponseWriter struct {
	*httptest.ResponseRecorder
	context.Context
}

type TestHandler string

func (h TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.(*TestResponseWriter).Context = r.Context()
	fmt.Fprint(w, string(h))
}

type ServeHTTPTest struct {
	Method string
	Path   string

	Status    int
	Body      string
	Variables []*Variable
	Header    http.Header
}

func TestMux_Handle_panicsWithBadRoutingConfiguration(t *testing.T) {
}

func TestMux_ServeHTTP_ServesAllRoutesWithoutTrailingCorrectly(t *testing.T) {
	m := New() //we are using the default error handlers

	m.Handle("/", TestHandler("ROOT"))

	catchall := m.SubRoute("/catchall")
	catchall.SubRoute("/*catchallvalue").Handle(TestHandler("CATCH_ALL"))
	catchall.SubRoute("/other").Handle(TestHandler("CATCH_ALL_OTHER"))
	catchall.SubRoute("/other/*catchallagain").Handle(TestHandler("CATCH_ALL_OTHER_AGAIN"))

	m.SubRoute("a").Handle(TestHandler("a"))
	m.SubRoute("a/b").Handle(TestHandler("b"))

	m.SubRoute("/c/").Handle(TestHandler("c/"))

	// t.Log(m.SubRoute("/c").node)

	//methods
	//segment variable with prefix
	// - followed by slash
	// - followed by not slash
	//end variable with prefix
	// - followed by slash
	// - followed by not slash
	//segment variable without prefix
	// - followed by slash
	// - followed by not slash
	//end variable without prefix
	// - followed by slash
	// - followed by not slash

	tests := []*ServeHTTPTest{
		{
			Method: "GET",
			Path:   "/",
			Status: 200,
			Body:   "ROOT",
		},
		{
			Method: "OPTIONS",
			Path:   "/",
			Status: 200,
			Body:   "ROOT",
		},
		{
			Method: "GET",
			Path:   "/notcatchall",
			Status: 404,
			Body:   NotFoundBody,
		},
		{
			Method: "GET",
			Path:   "/catchall",
			Status: 404,
			Body:   NotFoundBody,
		},
		{
			Method: "GET",
			Path:   "/catchall/will/catch/)(*$)(@&$_ANYTHING_IN_URL_PATH",
			Status: 200,
			Body:   "CATCH_ALL",
			Variables: []*Variable{
				{"catchallvalue", "will/catch/)(*$)(@&$_ANYTHING_IN_URL_PATH"},
			},
		},
		{
			Method: "POST",
			Path:   "/catchall/",
			Status: 200,
			Body:   "CATCH_ALL",
			Variables: []*Variable{
				{"catchallvalue", ""},
			},
		},
		{
			Method: "GET",
			Path:   "/catchall/oth",
			Status: 200,
			Body:   "CATCH_ALL",
			Variables: []*Variable{
				{"catchallvalue", "oth"},
			},
		},
		{
			Method: "GET",
			Path:   "/catchall/other",
			Status: 200,
			Body:   "CATCH_ALL_OTHER",
		},
		{
			Method: "GET",
			Path:   "/catchall/other/",
			Status: 200,
			Body:   "CATCH_ALL_OTHER_AGAIN",
			Variables: []*Variable{
				{"catchallagain", ""},
			},
		},
		{
			Method: "GET",
			Path:   "/catchall/other/again",
			Status: 200,
			Body:   "CATCH_ALL_OTHER_AGAIN",
			Variables: []*Variable{
				{"catchallagain", "again"},
			},
		},
	}

	testMux_ServeHTTP(t, m, tests...)
}

func TestMux_ServeHTTP_ServesAllRoutesWithAllowTrailingCorrectly(t *testing.T) {
}

func TestMux_ServeHTTP_ServesUnhandledRootWithANotFound(t *testing.T) {
	m := New()

	testMux_ServeHTTP(
		t,
		m,
		&ServeHTTPTest{
			Method: "GET",
			Path:   "",
			Status: 404,
			Body:   NotFoundBody,
		},
	)
}

func testMux_ServeHTTP(t *testing.T, m *Mux, tests ...*ServeHTTPTest) {
	for i, test := range tests {
		w := &TestResponseWriter{
			ResponseRecorder: httptest.NewRecorder(),
		}
		r, err := http.NewRequest(test.Method, test.Path, nil)
		if err != nil {
			t.Fatal(err)
		}
		w.Context = r.Context()

		m.ServeHTTP(w, r)

		if w.Code != test.Status {
			t.Errorf("%v: w.Status = %v WANT %v", i, w.Code, test.Status)
		}

		body := w.Body.String()
		if body != test.Body {
			t.Errorf("%v: w.Body = %v WANT %v", i, body, test.Body)
		}

		if vars := VariablesFrom(w.Context); !reflect.DeepEqual(vars, test.Variables) {
			t.Errorf("%v: result Variables = %v WANT %v", i, vars, test.Variables)
		}

		for vi, v := range test.Variables {
			actual, ok := VariableFromOk(w.Context, string(v.Name))
			if !ok || !reflect.DeepEqual(actual, v) {
				t.Errorf("%v: result Variable at %v = %v WANT %v", i, vi, actual, v)
			}
		}

		for key, values := range test.Header {
			if actual := w.Header().Get(key); !reflect.DeepEqual(actual, values) {
				t.Errorf("%v: desired Header %v = %v WANT %v", i, key, actual, values)
			}
		}
	}
}
