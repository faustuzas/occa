package errors

import "fmt"

type ErrorType int

const (
	Unauthorized ErrorType = iota + 1
	BadRequest
	InternalServer
)

func (t ErrorType) String() string {
	switch t {
	case Unauthorized:
		return "unauthorized"
	case BadRequest:
		return "bad_request"
	case InternalServer:
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
		type_: Unauthorized,
		cause: cause,
	}
}
