package errors

import "fmt"

type ErrorType int

const (
	TypeUnauthorized ErrorType = iota + 1
	TypeBadRequest
	TypeInternalServer
)

func (t ErrorType) String() string {
	switch t {
	case TypeUnauthorized:
		return "unauthorized"
	case TypeBadRequest:
		return "bad_request"
	case TypeInternalServer:
		return "internal"
	}
	panic(fmt.Sprintf("unrecognized error: %d", t))
}

type GenericErr struct {
	type_ ErrorType
	cause error
}

func (e GenericErr) Error() string {
	return fmt.Sprintf("[%s]: %v", e.type_, e.cause.Error())
}

func (e GenericErr) Type() ErrorType {
	return e.type_
}

func (e GenericErr) Unwrap() error {
	return e.cause
}

func ErrUnauthorized(cause error) GenericErr {
	return GenericErr{
		type_: TypeUnauthorized,
		cause: cause,
	}
}

func BadRequest(cause error) GenericErr {
	return GenericErr{
		type_: TypeBadRequest,
		cause: cause,
	}
}
