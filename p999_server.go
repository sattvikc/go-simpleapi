package fastapi

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	mux         *httprouter.Router
	swaggerJson map[string]interface{}
}

func New() *Server {
	s := &Server{
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
	addSwaggerRoutes(s)
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) GET(path string, handlers ...interface{}) error {
	handlerInstances, err := getHandlerInstances(handlers...)
	if err != nil {
		return err
	}

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

func (s *Server) addToSwagger(path string, handlers []Handler, method string, tags []string, responseTypes []struct {
	code        int
	response    interface{}
	description string
}) {
	definition := map[string]interface{}{
		"parameters": []interface{}{},
		"responses":  map[string]interface{}{},
		"tags":       tags,
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

	responses := definition["responses"].(map[string]interface{})
	for _, responseType := range responseTypes {
		codeStr := fmt.Sprintf("%d", responseType.code)
		schema := getSwaggerSchemaFromType(reflect.TypeOf(responseType.response))

		if _, ok := responses[codeStr]; !ok {
			responses[codeStr] = map[string]interface{}{
				"description": responseType.description,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": schema,
					},
				},
			}
		} else {
			responses[codeStr].(map[string]interface{})["description"] = responses[codeStr].(map[string]interface{})["description"].(string) + " or " + responseType.description

			eSchema := responses[codeStr].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
			if _, ok := eSchema["oneOf"]; !ok {
				responses[codeStr].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"] = map[string]interface{}{
					"oneOf": []interface{}{
						eSchema,
						schema,
					},
				}
			} else {
				eSchema["oneOf"] = append(eSchema["oneOf"].([]interface{}), schema)
			}
		}
	}
}

func (s *Server) Endpoint(path string, handlerFuncs ...func(e *Endpoint) interface{}) {
	e := &Endpoint{
		server:   s,
		path:     path,
		handlers: make([]interface{}, len(handlerFuncs)),
		tags:     []string{},
	}

	for i, handlerFunc := range handlerFuncs {
		e.handlers[i] = handlerFunc(e)
	}

	handlerInstances, err := getHandlerInstances(e.handlers...)
	if err != nil {
		panic(err)
	}

	e.handlerInstances = handlerInstances
	e.register()
}
