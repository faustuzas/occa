package http

import (
	"errors"
	"fmt"
	"net/http"

	pkgerrors "github.com/faustuzas/occa/src/pkg/errors"
)

type Err struct {
	StatusCode int
	Details    string
}

func (e Err) Error() string {
	return fmt.Sprintf("HTTP error (%v): %v", e.StatusCode, e.Details)
}

func DetermineHTTPError(err error) Err {
	var (
		statusCode = http.StatusInternalServerError
		cause      = err
	)

	var gErr pkgerrors.GenericErr
	if errors.As(err, &gErr) {
		switch gErr.Type() {
		case pkgerrors.BadRequest:
			statusCode = http.StatusBadRequest
		case pkgerrors.Unauthorized:
			statusCode = http.StatusUnauthorized
		case pkgerrors.InternalServer:
			statusCode = http.StatusInternalServerError
		}
		cause = gErr.Unwrap()
	}

	return Err{
		StatusCode: statusCode,
		Details:    cause.Error(),
	}
}
