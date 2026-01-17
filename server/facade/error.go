package facade

import (
	"fmt"
	"runtime"

	"github.com/futurxlab/golanggraph/xerror"
)

var (
	ErrServerInternal = &Error{
		Code:          20001,
		HttpStatus:    500,
		FacadeMessage: "Server Internal Error",
	}
	ErrBadRequest = &Error{
		Code:          20002,
		HttpStatus:    400,
		FacadeMessage: "Bad Request",
	}
	ErrUnauthorized = &Error{
		Code:          20003,
		HttpStatus:    401,
		FacadeMessage: "Unauthorized",
	}
	ErrorNotFound = &Error{
		Code:          20004,
		HttpStatus:    404,
		FacadeMessage: "Not Found",
	}
	ErrForbidden = &Error{
		Code:          20005,
		HttpStatus:    403,
		FacadeMessage: "Forbidden",
	}
	ErrTooManyRequests = &Error{
		Code:          20006,
		HttpStatus:    429,
		FacadeMessage: "Too Many Requests",
	}
)

type Error struct {
	HttpStatus    int    `json:"-" swaggerignore:"true"`
	Code          int    `json:"code"`
	FacadeMessage string `json:"msg"`
	InternalError error  `json:"-" swaggerignore:"true"`
}

func (e *Error) StatusCode() int {
	return e.HttpStatus
}

func (e *Error) Wrap(err error) *Error {
	res := &Error{
		HttpStatus:    e.HttpStatus,
		Code:          e.Code,
		FacadeMessage: e.FacadeMessage,
		InternalError: e.InternalError,
	}

	if res.InternalError == nil {
		res.InternalError = fmt.Errorf("%w", xerror.Wrap(err))
	} else {
		res.InternalError = fmt.Errorf("%w\n%w", xerror.Wrap(err), res.InternalError)
	}
	return res
}

func (e *Error) Facade(errorMessage string, params ...any) *Error {
	res := &Error{
		HttpStatus:    e.HttpStatus,
		Code:          e.Code,
		FacadeMessage: e.FacadeMessage,
		InternalError: e.InternalError,
	}

	errorMessage = fmt.Sprintf(errorMessage, params...)
	res.FacadeMessage = fmt.Sprintf("%s: %s", e.FacadeMessage, errorMessage)

	caller := e.getCaller(2)
	if res.InternalError == nil {
		res.InternalError = fmt.Errorf("%s", caller)
	} else {
		res.InternalError = fmt.Errorf("%s\n%w", caller, res.InternalError)
	}
	return res
}

func (e *Error) getCaller(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s %d", file, line)
}

// Convenience functions for creating common errors
func NewBadRequestError(message string, details ...string) *Error {
	err := &Error{
		HttpStatus:    400,
		Code:          20002,
		FacadeMessage: message,
	}
	if len(details) > 0 {
		err.FacadeMessage = fmt.Sprintf("%s: %s", message, details[0])
	}
	return err
}

func NewInternalServerError(message string, details ...string) *Error {
	err := &Error{
		HttpStatus:    500,
		Code:          20001,
		FacadeMessage: message,
	}
	if len(details) > 0 {
		err.FacadeMessage = fmt.Sprintf("%s: %s", message, details[0])
	}
	return err
}

// NewNotFoundError creates a new not found error (404) with a custom message
// Optional details parameter appends additional context to the message
func NewNotFoundError(message string, details ...string) *Error {
	err := &Error{
		HttpStatus:    404,
		Code:          20004,
		FacadeMessage: message,
	}
	if len(details) > 0 {
		err.FacadeMessage = fmt.Sprintf("%s: %s", message, details[0])
	}
	return err
}

func NewForbiddenError(message string, details ...string) *Error {
	err := &Error{
		HttpStatus:    403,
		Code:          20005,
		FacadeMessage: message,
	}
	if len(details) > 0 {
		err.FacadeMessage = fmt.Sprintf("%s: %s", message, details[0])
	}
	return err
}

func NewUnauthorizedError(message string, details ...string) *Error {
	err := &Error{
		HttpStatus:    401,
		Code:          20003,
		FacadeMessage: message,
	}
	if len(details) > 0 {
		err.FacadeMessage = fmt.Sprintf("%s: %s", message, details[0])
	}
	return err
}
