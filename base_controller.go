package fastcontroller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type fnAction func(*Context) error

type Controller struct {
	Log    logrus.FieldLogger
	Config Config
}

func NewController(l logrus.FieldLogger, c Config) Controller {
	return Controller{l, c}
}

func (c Controller) Handle(f fnAction) fasthttp.RequestHandler {
	return func(req *fasthttp.RequestCtx) {
		beginTime := time.Now()
		ctx := &Context{RequestCtx: req}
		defer logRequest(*ctx, c.Log, beginTime)
		defer handlePanic(ctx, c.Log)

		if err := f(ctx); err != nil {
			handleHttpError(ctx, err)
		}
	}
}

func (c Controller) Authorize(f fnAction, r Role, perms ...Permission) fnAction {
	return func(ctx *Context) error {
		tkn := ctx.Request.Header.Cookie("access_token")
		if r == ServiceRole {
			tkn = ctx.Request.Header.Peek("Authorization")
		}

		if string(tkn) == "" && (r != NoRole || len(perms) > 0) {
			return ErrUnauthorized(errors.Wrap(errors.New("empty Authorization header or access_token cookie"), "invalid token"))
		}

		claims, err := GetClaimsFromJWT(c.Config.JWT, tkn)
		if err != nil {
			if r == ServiceRole {
				return err
			}

			// TODO: refresh token for users

			return err
		}

		if r != NoRole && claims.Role != r {
			return ErrForbiden()
		}

		if len(perms) > 0 {
			for _, p := range perms {
				if !PermissionExist(p, claims.Permissions) {
					return ErrForbiden()
				}
			}
		}

		if r != ServiceRole {
			ctx = ctx.WithIdentify(claims.ID, claims.Username, claims.Role, claims.Permissions...)
		}

		return f(ctx)
	}
}

func (c Controller) SetJWT(ctx *Context, tkn string) {
	cookie := new(fasthttp.Cookie)
	cookie.SetKey("access_token")
	cookie.SetValue("Bearer " + tkn)
	cookie.SetMaxAge(int(c.Config.JWT.MaxAge))
	cookie.SetPath(c.Config.JWT.Path)
	cookie.SetSecure(c.Config.JWT.Secure)

	ctx.Response.Header.SetCookie(cookie)
}

func (c Controller) EncodeJson(ctx *Context, statusCode int, v interface{}) error {
	if v != nil {
		if err := c.jsonToResponse(&ctx.Response, v); err != nil {
			return errors.Wrap(err, "")
		}
	}

	ctx.Response.SetStatusCode(statusCode)

	return nil
}

func (c Controller) Response(ctx *Context, statusCode int) error {
	ctx.Response.SetStatusCode(statusCode)

	return nil
}

func (c Controller) DecodeJson(ctx *Context, v interface{}) error {
	err := json.Unmarshal(ctx.PostBody(), &v)
	if err != nil {
		return ErrValidation(err.Error(), errors.Wrap(err, ""))
	}

	return nil
}

func (c Controller) View(ctx *Context, buff *bytes.Buffer) {
	ctx.Response.Header.Add("Content-Type", "text/html")
	ctx.Response.SetStatusCode(http.StatusOK)
	ctx.Response.AppendBodyString(buff.String())
}

func (c Controller) jsonToResponse(r *fasthttp.Response, v interface{}) error {
	if err := json.NewEncoder(r.BodyWriter()).Encode(&v); err != nil {
		return errors.Wrap(err, "")
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("charset", "utf8")

	return nil
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
