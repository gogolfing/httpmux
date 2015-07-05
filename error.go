package mux

import (
	"net/http"
	"strings"
)

const (
	headerAllow = "Allow"
	ErrNotFound = ErrStatusHandler(http.StatusNotFound)
)

type ErrStatusHandler int

func (h ErrStatusHandler) Error() string {
	return http.StatusText(int(h))
}

func (h ErrStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveErrorStatus(w, int(h))
}

type ErrMethodNotAllowed []string

func (_ ErrMethodNotAllowed) Error() string {
	return http.StatusText(http.StatusMethodNotAllowed)
}

func (e ErrMethodNotAllowed) Header() string {
	return strings.Join(e, ", ")
}

func (e ErrMethodNotAllowed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headerAllow, e.Header())
	serveErrorStatus(w, http.StatusMethodNotAllowed)
}

func serveErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
