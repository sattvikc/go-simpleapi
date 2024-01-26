package fastapi

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/sattvikc/go-strapi/handler"
	"github.com/sattvikc/go-strapi/router"
	"github.com/sattvikc/go-strapi/swagger"
)

type App struct {
	r           *router.Router
	swaggerJson map[string]interface{}
}

func New() *App {
	s := &App{
		r: router.New(),
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

func (s *App) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, params := s.r.FindCall(r.URL.Path, r.Method)

	if h == nil {
		// TODO handler not found
		return
	}

	ctx := &Context{
		Request:  r,
		Response: w,
		params:   params,
		next:     h.(*handler.Handler).Clone(),
	}

	err := ctx.Next()
	if err != nil {
		fmt.Println(err)
		// TODO handle error
	}
}

func (s *App) AddHandler(path string, method string, handlers ...interface{}) error {
	h, err := handler.New(handlers...)
	if err != nil {
		return err
	}
	s.r.Add(path, method, h, "")

	return nil
}

func (s *App) addToSwagger(path string, handlers *handler.Handler, method string, tags []string, responseTypes []struct {
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

	for _, handler := range *handlers {
		swagger.UpdateDefinitionUsingParamTypes(definition, handler.ParamTypes)
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
		schema := swagger.GetSwaggerSchemaForType(reflect.TypeOf(responseType.response))

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

func (s *App) Endpoint(path string, handlerFuncs ...func(e *Endpoint) interface{}) {
	e := &Endpoint{
		app:      s,
		path:     path,
		handlers: make([]interface{}, len(handlerFuncs)),
		tags:     []string{},
	}

	for i, handlerFunc := range handlerFuncs {
		e.handlers[i] = handlerFunc(e)
	}

	handlerInstances, err := handler.New(e.handlers...)
	if err != nil {
		panic(err)
	}

	e.handlerInstances = handlerInstances
	e.register()
}
