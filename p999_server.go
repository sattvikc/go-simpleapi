package fastapi

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	mux         *httprouter.Router
	swaggerJson map[string]interface{}
}

func New() *Server {
	return &Server{
		mux: httprouter.New(),
		swaggerJson: map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "FastAPI",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{},
		},
	}
}

func (s *Server) ListenAndServe(addr string) error {
	addSwaggerRoutes(s)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) GET(path string, handlers ...interface{}) error {
	handlerInstances, err := getHandlerInstances(handlers...)
	if err != nil {
		return err
	}

	s.addToSwagger(path, handlerInstances, "get")

	s.mux.GET(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{
			Request:      r,
			Response:     w,
			params:       p,
			nextHandlers: handlerInstances,
		}
		err := ctx.Next()
		if err != nil {
			fmt.Println(err)
			// TODO handle error
		}
	})

	return nil
}

func (s *Server) POST(path string, handlers ...interface{}) error {

	handlerInstances, err := getHandlerInstances(handlers...)
	if err != nil {
		return err
	}

	s.addToSwagger(path, handlerInstances, "post")

	s.mux.POST(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{
			Request:      r,
			Response:     w,
			params:       p,
			nextHandlers: handlerInstances,
		}
		err := ctx.Next()
		if err != nil {
			fmt.Println(err)
			// TODO handle error
		}
	})

	return nil
}

func (s *Server) PUT(path string, handlers ...interface{}) error {

	handlerInstances, err := getHandlerInstances(handlers...)
	if err != nil {
		return err
	}

	s.addToSwagger(path, handlerInstances, "put")

	s.mux.PUT(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{
			Request:      r,
			Response:     w,
			params:       p,
			nextHandlers: handlerInstances,
		}
		err := ctx.Next()
		if err != nil {
			fmt.Println(err)
			// TODO handle error
		}
	})

	return nil
}

func (s *Server) DELETE(path string, handlers ...interface{}) error {

	handlerInstances, err := getHandlerInstances(handlers...)
	if err != nil {
		return err
	}

	s.addToSwagger(path, handlerInstances, "delete")

	s.mux.DELETE(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{
			Request:      r,
			Response:     w,
			params:       p,
			nextHandlers: handlerInstances,
		}
		err := ctx.Next()
		if err != nil {
			fmt.Println(err)
			// TODO handle error
		}
	})

	return nil
}

func (s *Server) addToSwagger(path string, handlers []Handler, method string) {
	definition := map[string]interface{}{
		"parameters": []interface{}{},
		"responses":  map[string]interface{}{},
	}

	if method != "get" {
		definition["requestBody"] = map[string]interface{}{}
	}

	for _, handler := range handlers {
		updateDefinitionFromhandler(definition, handler.paramTypes)
	}

	if _, ok := s.swaggerJson["paths"].(map[string]interface{})[path]; !ok {
		s.swaggerJson["paths"].(map[string]interface{})[path] = map[string]interface{}{}
	}

	if _, ok := s.swaggerJson["paths"].(map[string]interface{})[path].(map[string]interface{})[method]; !ok {
		s.swaggerJson["paths"].(map[string]interface{})[path].(map[string]interface{})[method] = definition
	}
}
