package errror

import (
	"fmt"
)

// we WRAP the error exactly where it appears and then we pass it up
// if after its transported up it need to be sent to admin, we ErrInJson it

func WrapErrorF(orig error, code ErrorCode, cause string, a ...interface{}) error {
	return &Error{
		Orig: orig,
		code: code,
		msg:  fmt.Sprintf(cause, a...),
	}
}

// Error returns the message, when wrapping errors the wrapped error is returned.
func (e *Error) Error() string {
	if e.Orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.Orig)
	}

	return e.msg
}

// NewErrorf instantiates a new error.
func NewErrorf(code ErrorCode, cause string, a ...interface{}) error {
	return WrapErrorF(nil, code, cause, a...)
}

// Code returns the code representing this error.
func (e *Error) Code() ErrorCode {
	return e.code
}

type Error struct {
	Orig error
	msg  string
	code ErrorCode
}

type ErrorCode uint

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeFailure
	ErrorCodeNotFound
	ErrorCodeInvalidArgument
	ErrorCodeIsBanned
)

// for not found and invalid argument, return to user
// for failure and (unknown) return to admin

// in each package errors can be handled differently
