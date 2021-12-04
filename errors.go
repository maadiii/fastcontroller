package fast

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func handleHttpError(ctx *Context, err error) {
	ctx.Response.Header.Add("Content-Type", "text/plain; charset=utf-8")
	ctx.Response.Header.Add("X-Content-Type-Options", "nosniff")
	if e, ok := err.(ErrorResponseType); ok {
		logrus.Errorf("%+v\r\n", e.InternalError())
		ctx.Response.SetStatusCode(e.HTTPCode())
		ctx.Response.SetBody([]byte(e.HTTPMessage()))
		return
	}

	logrus.Errorf("%+v\r\n", err)
	ctx.Response.SetStatusCode(http.StatusInternalServerError)
	ctx.Response.SetBody([]byte(http.StatusText(http.StatusInternalServerError)))
}

type ErrorResponseType interface {
	HTTPCode() int
	HTTPMessage() string
	InternalError() error
}

type errorResponseType struct {
	httpCode      int
	httpMessage   string
	internalError error
}

func (e errorResponseType) HTTPCode() int {
	return e.httpCode
}

func (e errorResponseType) HTTPMessage() string {
	return e.httpMessage
}

func (e errorResponseType) InternalError() error {
	return e.internalError
}

func (e errorResponseType) Error() string {
	return e.InternalError().Error()
}

func ErrNotImplemented(fnName string) errorResponseType {
	return errorResponseType{
		httpCode:      http.StatusNotImplemented,
		httpMessage:   http.StatusText(http.StatusNotImplemented),
		internalError: errors.New(fmt.Sprintf("%s Not Implemented", fnName)),
	}
}

func ErrNotFound(field string, err error) errorResponseType {
	return errorResponseType{
		httpCode:      http.StatusNotFound,
		httpMessage:   fmt.Sprintf("%s %s", field, http.StatusText(http.StatusNotFound)),
		internalError: err,
	}
}

func ErrUnauthorized(err error) errorResponseType {
	return errorResponseType{
		httpCode:      http.StatusUnauthorized,
		httpMessage:   http.StatusText(http.StatusUnauthorized),
		internalError: err,
	}
}

func ErrForbiden() errorResponseType {
	return errorResponseType{
		httpCode:    http.StatusForbidden,
		httpMessage: http.StatusText(http.StatusForbidden),
	}
}

func ErrValidation(message string, err error) errorResponseType {
	return errorResponseType{
		httpCode:      http.StatusBadRequest,
		httpMessage:   message,
		internalError: err,
	}
}
