package fastapi

import (
	"errors"
	"reflect"
)

type Handler struct {
	handlerFunc reflect.Value
	paramTypes  []reflect.Type
}

func (h Handler) handle(ctx *Context) error {
	params := make([]reflect.Value, len(h.paramTypes))

	for idx, paramType := range h.paramTypes {
		param := reflect.New(paramType).Elem()

		err := populateValueFromTypeUsingContext(ctx, paramType, param)
		if err != nil {
			return err
		}
		params[idx] = param
	}

	result := h.handlerFunc.Call(append([]reflect.Value{
		reflect.ValueOf(ctx),
	}, params...))

	res := result[0].Interface()
	if res != nil {
		return res.(error)
	}

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

		if numParams == 0 {
			return nil, errors.New("handler must have at least 1 parameter")
		}

		paramTypes := make([]reflect.Type, numParams-1)

		for i := 1; i < numParams; i++ {
			paramTypes[i-1] = handlerFunc.In(i)

			if paramTypes[i-1].Kind() != reflect.Struct {
				return nil, errors.New("handler parameter must be a struct")
			}
		}

		handlerInstances[idx].handlerFunc = funcValue
		handlerInstances[idx].paramTypes = paramTypes
	}

	return handlerInstances, nil
}
