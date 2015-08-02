package httpmux

import (
	"net/http"
	"testing"
)

type testRoute struct {
	method string
	path   string
}

type emptyResponseWriter struct{}

func (_ *emptyResponseWriter) Header() http.Header {
	return http.Header{}
}

func (_ *emptyResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (_ *emptyResponseWriter) WriteHeader(code int) {
}

type emptyHandler struct{}

func (_ *emptyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (m *Mux) handleTestRoutes(routes []testRoute, handler http.Handler) {
	for _, route := range routes {
		m.Handle(route.path, handler, route.method)
	}
}

func benchmarkRoutes(b *testing.B, m *Mux, routes []testRoute) {
	w := &emptyResponseWriter{}
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			m.ServeHTTP(w, r)
		}
	}
}
