package mux

import "testing"

func TestMux(t *testing.T) {
	m := New()
	m.Handle("/", intHandler(0))
}
