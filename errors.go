package httpmux

import (
	"fmt"
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

type ErrOverlapStaticVar struct {
	StaticPath string
	Variable   string
}

func (e *ErrOverlapStaticVar) Error() string {
	return fmt.Sprintf("cannot have static path: %q and variable %q at the same location", e.StaticPath, e.Variable)
}

type ErrConsecutiveVars struct {
	Variable1 string
	Variable2 string
}

func (e *ErrConsecutiveVars) Error() string {
	return fmt.Sprintf("cannot have two immediately consecutive variables: %q, %q", e.Variable1, e.Variable2)
}

type ErrUnequalVars struct {
	Variable1 string
	Variable2 string
}

func (e *ErrUnequalVars) Error() string {
	return fmt.Sprintf("cannot have two unequal variables at the same location: %q, %q", e.Variable1, e.Variable2)
}
