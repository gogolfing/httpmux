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
			Status: 404,
			Body:   NotFoundBody,
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
			Body:   "CATCH_ALL",
			Variables: []*Variable{
				{"catchallvalue", "other/"},
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

	testMux_ServeHTTP_result(t, m, tests...)
}

func TestMux_ServeHTTP_ServesAllRoutesWithAllowTrailingCorrectly(t *testing.T) {
}

func TestMux_ServeHTTP_ServesUnhandledRootWithANotFound(t *testing.T) {
	m := New()

	testMux_ServeHTTP_result(
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

func testMux_ServeHTTP_result(t *testing.T, m *Mux, tests ...*ServeHTTPTest) {
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

		t.Log("testMux_ServeHTTP_result", i, r.URL.Path)

		if w.Code != test.Status {
			t.Errorf("w.Status = %v WANT %v", w.Code, test.Status)
		}

		body := w.Body.String()
		if body != test.Body {
			t.Errorf("w.Body = %v WANT %v", body, test.Body)
		}

		if vars := VariablesFrom(w.Context); !reflect.DeepEqual(vars, test.Variables) {
			t.Errorf("result Variables = %v WANT %v", vars, test.Variables)
		}

		for i, v := range test.Variables {
			actual, ok := VariableFromOk(w.Context, string(v.Name))
			if !ok || !reflect.DeepEqual(actual, v) {
				t.Errorf("result Variable at %v = %v WANT %v", i, actual, v)
			}
		}

		for key, values := range test.Header {
			if actual := w.Header().Get(key); !reflect.DeepEqual(actual, values) {
				t.Errorf("desired Header %v = %v WANT %v", key, actual, values)
			}
		}
	}
}
