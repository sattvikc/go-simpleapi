package simpleapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sattvikc/go-simpleapi/handler"
	"github.com/sattvikc/go-simpleapi/router"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	params   router.Params
	next     *handler.Handler
}

func (c *Context) Next() error {
	if !c.next.HasNext() {
		return nil
	}

	nextHandler := c.next.Get()

	return nextHandler.Invoke(c, c.Request, c.params)
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
