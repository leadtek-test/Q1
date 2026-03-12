package errors

import (
	"errors"
	"fmt"

	"github.com/leadtek-test/q1/common/consts"
)

type Error struct {
	code       int
	msg        string
	err        error
	statusCode int
}

func (e *Error) Error() string {
	var msg string
	msg = consts.ErrMsg[e.code]
	if e.msg != "" {
		msg = e.msg
	}
	if e.err == nil {
		return msg
	}
	return msg + " -> " + e.err.Error()
}

func New(code int) error {
	return &Error{
		code:       code,
		statusCode: consts.StatusCodeByErrno(code),
	}
}

func NewWithError(code int, err error) error {
	if err == nil {
		return New(code)
	}
	return &Error{
		code:       code,
		err:        err,
		statusCode: consts.StatusCodeByErrno(code),
	}
}

func NewWithMsgf(code int, format string, args ...any) error {
	return &Error{
		code:       code,
		msg:        fmt.Sprintf(format, args...),
		statusCode: consts.StatusCodeByErrno(code),
	}
}

func NewWithStatusCode(code int, statusCode int) error {
	return &Error{
		code:       code,
		statusCode: normalizeStatusCode(code, statusCode),
	}
}

func NewWithErrorAndStatusCode(code int, statusCode int, err error) error {
	if err == nil {
		return NewWithStatusCode(code, statusCode)
	}
	return &Error{
		code:       code,
		err:        err,
		statusCode: normalizeStatusCode(code, statusCode),
	}
}

func NewWithMsgfAndStatusCode(code int, statusCode int, format string, args ...any) error {
	return &Error{
		code:       code,
		msg:        fmt.Sprintf(format, args...),
		statusCode: normalizeStatusCode(code, statusCode),
	}
}

func Errno(err error) int {
	if err == nil {
		return consts.ErrnoSuccess
	}
	if targetError, ok := errors.AsType[*Error](err); ok {
		return targetError.code
	}
	return -1
}

func Output(err error) (int, string) {
	if err == nil {
		return consts.ErrnoSuccess, consts.ErrMsg[consts.ErrnoSuccess]
	}
	errno := Errno(err)
	if errno == -1 {
		return consts.ErrnoUnknownError, err.Error()
	}
	return errno, err.Error()
}

func StatusCode(err error) int {
	if err == nil {
		return consts.StatusCodeByErrno(consts.ErrnoSuccess)
	}

	if targetError, ok := errors.AsType[*Error](err); ok {
		if targetError.statusCode != 0 {
			return targetError.statusCode
		}
		return consts.StatusCodeByErrno(targetError.code)
	}
	return consts.StatusCodeByErrno(consts.ErrnoUnknownError)
}

func OutputWithStatus(err error) (int, string, int) {
	errno, msg := Output(err)
	return errno, msg, StatusCode(err)
}

func normalizeStatusCode(code int, statusCode int) int {
	if statusCode < 100 || statusCode > 999 {
		return consts.StatusCodeByErrno(code)
	}
	return statusCode
}
