package fast

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	fnAction       func(*Context) error
	fnAuthorzie    func(*Context, config.JWT, Role, ...Permission) error
	fnRefreshToken func(id int64) (string, error)
)

func (a fnAction) Invoke(ctx *Context) {
	if err := a(ctx); err != nil {
		handleHttpError(ctx, err)
	}
}

func handlePanic(ctx *Context, log logrus.FieldLogger) {
	if r := recover(); r != nil {
		log.Errorf("%v: %s", r, debug.Stack())
		ctx.Response.Header.Add("Content-Type", "text/plain; charset=utf-8")
		ctx.Response.Header.Add("X-Content-Type-Options", "nosniff")
		ctx.Response.SetStatusCode(http.StatusInternalServerError)
		ctx.Response.SetBody([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}

func logRequest(ctx Context, l logrus.FieldLogger, beginTime time.Time) {
	logger := l.WithFields(
		logrus.Fields{
			"duration":    time.Since(beginTime),
			"status_code": ctx.Response.StatusCode,
			"remote":      ctx.ReadUserIP(),
			"status":      ctx.Response.StatusCode(),
		},
	)
	logger.Info(
		string(ctx.RequestCtx.Method()), string(ctx.Request.URI().RequestURI()),
	)
}
