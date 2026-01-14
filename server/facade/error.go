package facade

import (
	"fmt"
	"runtime"

	"github.com/futurxlab/golanggraph/xerror"
)

var (
	ErrServerInternal = &FuturxError{
		Code:          20001,
		HttpStatus:    500,
		FacadeMessage: "Server Internal Error",
	}
	ErrBadRequest = &FuturxError{
		Code:          20002,
		HttpStatus:    400,
		FacadeMessage: "Bad Request",
	}
	ErrUnauthorized = &FuturxError{
		Code:          20003,
		HttpStatus:    401,
		FacadeMessage: "Unauthorized",
	}
	ErrorNotFound = &FuturxError{
		Code:          20004,
		HttpStatus:    404,
		FacadeMessage: "Not Found",
	}
	ErrForbidden = &FuturxError{
		Code:          20005,
		HttpStatus:    403,
		FacadeMessage: "Forbidden",
	}
	ErrTooManyRequests = &FuturxError{
		Code:          20006,
		HttpStatus:    429,
		FacadeMessage: "Too Many Requests",
	}
)

type FuturxError struct {
	HttpStatus    int    `json:"-" swaggerignore:"true"`
	Code          int    `json:"code"`
	FacadeMessage string `json:"msg"`
	InternalError error  `json:"-" swaggerignore:"true"`
}

func (e *FuturxError) StatusCode() int {
	return e.HttpStatus
}

func (e *FuturxError) Wrap(err error) *FuturxError {
	res := &FuturxError{
		HttpStatus:    e.HttpStatus,
		Code:          e.Code,
		FacadeMessage: e.FacadeMessage,
		InternalError: e.InternalError,
	}

	if res.InternalError == nil {
		res.InternalError = fmt.Errorf("%w", xerror.WrapWithCaller(err, 2))
	} else {
		res.InternalError = fmt.Errorf("%w\n%w", xerror.WrapWithCaller(err, 2), res.InternalError)
	}
	return res
}

func (e *FuturxError) Facade(errorMessage string, params ...any) *FuturxError {
	res := &FuturxError{
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

func (e *FuturxError) getCaller(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s %d", file, line)
}
