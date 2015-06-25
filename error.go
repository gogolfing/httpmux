package mux

import (
	"errors"
	"net/http"
	"strings"
)

const (
	HeaderAllow = "Allow"
)

var (
	ErrNotFound = errors.New(http.StatusText(http.StatusNotFound))
)

type ErrStatusHandler int

func (h ErrStatusHandler) Error() string {
	return http.StatusText(int(h))
}

func (h ErrStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveErrorStatus(w, int(h))
}

type ErrMethodNotAllowed struct {
	methods []string
}

func (_ *ErrMethodNotAllowed) Error() string {
	return http.StatusText(http.StatusMethodNotAllowed)
}

func (e *ErrMethodNotAllowed) Header() string {
	return strings.Join(e.methods, ", ")
}

func (e *ErrMethodNotAllowed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(HeaderAllow, e.Header())
	serveErrorStatus(w, http.StatusMethodNotAllowed)
}

func serveErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
