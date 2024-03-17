package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgerrors "github.com/faustuzas/occa/src/pkg/errors"
)

func Error(err error) error {
	var (
		statusCode = codes.Internal
		cause      = err
	)

	var gErr pkgerrors.GenericErr
	if errors.As(err, &gErr) {
		switch gErr.Type() {
		case pkgerrors.TypeBadRequest:
			statusCode = codes.InvalidArgument
		case pkgerrors.TypeUnauthorized:
			statusCode = codes.Unauthenticated
		case pkgerrors.TypeInternalServer:
			statusCode = codes.Internal
		}
		cause = gErr.Unwrap()
	}

	return status.Error(statusCode, cause.Error())
}
