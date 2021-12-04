package fast

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/valyala/fasthttp"
)

func SendRequest(method, uri string, body []byte, response interface{}) error {
	if reflect.ValueOf(response).Kind() != reflect.Ptr {
		return errors.New("response type must be a pointer value")
	}

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(method)
	req.Header.SetContentType("application/json")
	if body != nil {
		req.SetBody(body)
	}
	req.SetRequestURI(uri)

	res := fasthttp.AcquireResponse()

	if err := fasthttp.Do(req, res); err != nil {
		return err
	}
	fasthttp.ReleaseRequest(req)

	if err := json.Unmarshal(res.Body(), response); err != nil {
		return err
	}
	fasthttp.ReleaseResponse(res)

	return nil
}
