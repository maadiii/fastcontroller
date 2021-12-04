package main

import "github.com/valyala/fasthttp"

type Context struct {
	*fasthttp.RequestCtx
	Identify *Identify
}

func (c *Context) WithIdentify(id uint, username string, r Role, p ...Permission) *Context {
	c.Identify = &Identify{Id: id, Username: username, Role: r, Permissions: p}

	return c
}

func (c Context) ReadUserIP() string {
	ip := string(c.Request.Header.Peek("X-Real-Ip"))
	if ip == "" {
		ip = string(c.Request.Header.Peek("X-Forwarded-For"))
	}

	return ip
}

type Identify struct {
	Id          uint
	Username    string
	Role        Role
	Permissions []Permission
}
