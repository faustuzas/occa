package http

import (
	"fmt"
	"net/http"
)

type Err struct {
	code  int
	cause error
}

func (e Err) Error() string {
	return fmt.Sprintf("HTTP error (%v): %v", e.code, e.cause.Error())
}

func ErrUnauthorized(err error) error {
	return Err{
		code:  http.StatusUnauthorized,
		cause: err,
	}
}
