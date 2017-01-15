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

type ErrOverlapStaticVar VarName

func (e ErrOverlapStaticVar) Error() string {
	return fmt.Sprintf("httpmux: cannot have static path and variable %q at the same location", string(e))
}

type ErrConsecutiveVars struct {
	Variable1 VarName
	Variable2 VarName
}

func (e *ErrConsecutiveVars) Error() string {
	return fmt.Sprintf("httpmux: cannot have two consecutive variables %q and %q", e.Variable1, e.Variable2)
}

type ErrUnequalVars struct {
	Variable1 VarName
	Variable2 VarName
}

func (e *ErrUnequalVars) Error() string {
	return fmt.Sprintf("httpmux: cannot have two unequal variables at the same location %q and %q", e.Variable1, e.Variable2)
}

type errSomethingAfterEndVar string

func (e errSomethingAfterEndVar) Error() string {
	return string(e)
}
