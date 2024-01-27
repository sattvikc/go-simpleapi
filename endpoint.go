package simpleapi

import "github.com/sattvikc/go-simpleapi/handler"

type Endpoint struct {
	app              *App
	path             string
	method           string
	handlers         []interface{}
	handlerInstances *handler.Handler
	tags             []string
	responseTypes    []struct {
		code        int
		response    interface{}
		description string
	}
}

func (e *Endpoint) WithTag(tag string) *Endpoint {
	e.tags = append(e.tags, tag)
	return e
}

func (e *Endpoint) WithResponse(code int, response interface{}, description string) *Endpoint {
	e.responseTypes = append(e.responseTypes, struct {
		code        int
		response    interface{}
		description string
	}{code: code, response: response, description: description})
	return e
}

func (e *Endpoint) GET() *Endpoint {
	e.method = "get"
	return e
}

func (e *Endpoint) POST() *Endpoint {
	e.method = "post"
	return e
}

func (e *Endpoint) PUT() *Endpoint {
	e.method = "put"
	return e
}

func (e *Endpoint) DELETE() *Endpoint {
	e.method = "delete"
	return e
}

func (e *Endpoint) register() error {
	e.app.addToSwagger(e.path, e.handlerInstances, e.method, e.tags, e.responseTypes)

	err := e.app.AddHandler(e.path, e.method, e.handlers...)
	if err != nil {
		return err
	}

	return nil
}
