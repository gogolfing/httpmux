package golfmux

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

func (h ErrStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeErrorStatus(w, int(h))
}

type ErrMethodNotAllowed struct {
	methods []string
}

func (_ *ErrMethodNotAllowed) Error() string {
	return http.StatusText(http.StatusMethodNotAllowed)
}

func (e *ErrMethodNotAllowed) header() string {
	return strings.Join(e.methods, ", ")
}

func (e *ErrMethodNotAllowed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(HeaderAllow, e.header())
	ServeErrorStatus(w, http.StatusMethodNotAllowed)
}

func ServeErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
