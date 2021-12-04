package fastcontroller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Controller struct {
	Log    logrus.FieldLogger
	Config *Config

	Authorize    fnAuthorzie
	RefreshToken fnRefreshToken
}

func NewController(l logrus.FieldLogger, a fnAuthorzie, r fnRefreshToken, c *Config) Controller {
	return Controller{l, c, a, r}
}

func (c *Controller) Handle(f fnAction, r Role, p ...Permission) fasthttp.RequestHandler {
	return func(req *fasthttp.RequestCtx) {
		beginTime := time.Now()
		ctx := &Context{RequestCtx: req}
		defer logRequest(*ctx, c.Log, beginTime)
		defer handlePanic(ctx, c.Log)

		if err := c.Authorize(ctx, c.Config.JWT, r, p...); err != nil {
			handleHttpError(ctx, err)
			return
		}

		f.Invoke(ctx)
	}
}

func (c *Controller) SetJWT(ctx *Context, tkn string) {
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
