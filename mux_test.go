package mux

import (
	"net/http"
	"net/http/httptest"
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
