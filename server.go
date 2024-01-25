package fastapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

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

	s.mux.GET("/openapi.json", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		swaggerBytes, err := json.Marshal(s.swaggerJson)
		if err != nil {
			return
		}
		w.Write(swaggerBytes)
	})

	s.mux.GET("/docs", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerUI))
	})

	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) POST(path string, handlers ...interface{}) error {

	handlerInstances, err := s.getHandlerInstances(handlers...)
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

func (s *Server) getHandlerInstances(handlers ...interface{}) ([]Handler, error) {
	handlerInstances := make([]Handler, len(handlers))

	for idx, handler := range handlers {
		handlerFunc := reflect.TypeOf(handler)

		if handlerFunc.Kind() != reflect.Func {
			return nil, errors.New("handler is not a function")
		}

		funcValue := reflect.ValueOf(handler)

		numParams := handlerFunc.NumIn()

		if numParams != 2 {
			return nil, errors.New("handler does not have 2 parameters")
		}

		paramType := handlerFunc.In(1)

		if paramType.Kind() != reflect.Struct {
			return nil, errors.New("handler does not have a struct as second parameter")
		}

		handlerInstances[idx].handlerFunc = funcValue
		handlerInstances[idx].paramType = paramType
	}

	return handlerInstances, nil
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
		s.updateDefinitionFromhandler(definition, handler.paramType)
	}

	if _, ok := s.swaggerJson["paths"].(map[string]interface{})[path]; !ok {
		s.swaggerJson["paths"].(map[string]interface{})[path] = map[string]interface{}{}
	}

	if _, ok := s.swaggerJson["paths"].(map[string]interface{})[path].(map[string]interface{})[method]; !ok {
		s.swaggerJson["paths"].(map[string]interface{})[path].(map[string]interface{})[method] = definition
	}
}

func (s *Server) updateDefinitionFromhandler(definition map[string]interface{}, paramType reflect.Type) {
	HEADER_EXCLUSIONS := map[string]bool{"content-type": true, "content-length": true, "user-agent": true}

	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		if field.Tag.Get("body") != "" {
			if field.Tag.Get("body") == "multipart" {
				properties := map[string]interface{}{}

				for j := 0; j < field.Type.NumField(); j++ {
					field := field.Type.Field(j)
					if field.Tag.Get("form") != "" {
						properties[field.Tag.Get("form")] = s.getSwaggerSchemaFromType(field.Type)
					}
				}

				bodyDefinition := map[string]interface{}{
					"content": map[string]interface{}{
						"multipart/form-data": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":       "object",
								"properties": properties,
							},
						},
					},
				}

				definition["requestBody"] = bodyDefinition

			} else if field.Tag.Get("body") == "urlencoded" {
				properties := map[string]interface{}{}

				for j := 0; j < field.Type.NumField(); j++ {
					field := field.Type.Field(j)
					if field.Tag.Get("form") != "" {
						properties[field.Tag.Get("form")] = s.getSwaggerSchemaFromType(field.Type)
					}
				}

				bodyDefinition := map[string]interface{}{
					"content": map[string]interface{}{
						"application/x-www-form-urlencoded": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":       "object",
								"properties": properties,
							},
						},
					},
				}

				definition["requestBody"] = bodyDefinition
			} else if field.Tag.Get("body") == "json" {
				bodyDefinition := map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": s.getSwaggerSchemaFromType(field.Type),
						},
					},
				}
				definition["requestBody"] = bodyDefinition
			}

		} else if field.Type.Kind() == reflect.Struct {
			s.updateDefinitionFromhandler(definition, field.Type)
		} else if field.Tag.Get("query") != "" {
			queryDefinition := map[string]interface{}{
				"in":       "query",
				"name":     field.Tag.Get("query"),
				"required": true, // TODO
				"schema": map[string]interface{}{
					"type": "string",
				},
			}
			definition["parameters"] = append(definition["parameters"].([]interface{}), queryDefinition)

		} else if field.Tag.Get("header") != "" {
			if _, ok := HEADER_EXCLUSIONS[strings.ToLower(field.Tag.Get("header"))]; ok {
				continue
			}
			headerDefinition := map[string]interface{}{
				"in":       "header",
				"name":     field.Tag.Get("header"),
				"required": true, // TODO
				"schema": map[string]interface{}{
					"type": "string",
				},
			}
			definition["parameters"] = append(definition["parameters"].([]interface{}), headerDefinition)

		} else if field.Tag.Get("path") != "" {
			pathDefinition := map[string]interface{}{
				"in":       "path",
				"name":     field.Tag.Get("path"),
				"required": true, // TODO
				"schema": map[string]interface{}{
					"type": "string",
				},
			}
			definition["parameters"] = append(definition["parameters"].([]interface{}), pathDefinition)

		}
	}
}

func (s *Server) getSwaggerSchemaFromType(t reflect.Type) interface{} {
	fileType := reflect.TypeOf((*multipart.File)(nil)).Elem()
	if t.ConvertibleTo(fileType) {
		return map[string]interface{}{
			"type": "file",
		}
	}

	if t.Kind() == reflect.Struct {
		properties := map[string]interface{}{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Tag.Get("json") != "" {
				properties[field.Tag.Get("json")] = s.getSwaggerSchemaFromType(field.Type)
			}
		}

		return map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{
			"type": "string",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{
			"type": "number",
		}
	case reflect.Bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	default:
		return map[string]interface{}{
			"type": "string",
		}
	}
}
