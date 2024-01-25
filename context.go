package fastapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	Request      *http.Request
	Response     http.ResponseWriter
	params       httprouter.Params
	nextHandlers []Handler
}

func (c *Context) Next() error {
	if len(c.nextHandlers) == 0 {
		return nil
	}

	nextHandler := c.nextHandlers[0]
	c.nextHandlers = c.nextHandlers[1:]

	return nextHandler.handle(c)
}

func (c *Context) JSON(status int, data interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(status)
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.Response.Header().Set("Content-Length", fmt.Sprint(len(dataBytes)))
	c.Response.Write(dataBytes)
	return nil
}

func (c *Context) HTML(status int, html string) error {
	c.Response.Header().Set("Content-Type", "text/html")
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(html))
	return nil
}
