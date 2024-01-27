package handler

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/sattvikc/go-simpleapi/reflection"
	"github.com/sattvikc/go-simpleapi/router"
)

type handler struct {
	Func       reflect.Value
	ParamTypes []reflect.Type
}

type Handler []handler

func (h handler) Invoke(ctx interface{}, request *http.Request, params router.Params) error {
	fParams := make([]reflect.Value, len(h.ParamTypes))

	for idx, paramType := range h.ParamTypes {
		param := reflect.New(paramType).Elem()

		err := reflection.PopulateValueFromTypeUsingContext(request, params, paramType, param)
		if err != nil {
			return err
		}
		fParams[idx] = param
	}

	result := h.Func.Call(append([]reflect.Value{
		reflect.ValueOf(ctx),
	}, fParams...))

	res := result[0].Interface()
	if res != nil {
		return res.(error)
	}

	return nil
}

func (h *Handler) HasNext() bool {
	return len(*h) > 0
}

func (h *Handler) Get() handler {
	next := (*h)[0]
	*h = (*h)[1:]
	return next
}

func (h *Handler) Clone() *Handler {
	newHandler := make(Handler, len(*h))
	copy(newHandler, *h)
	return &newHandler
}

func New(handlers ...interface{}) (*Handler, error) {
	handlerInstances := make(Handler, len(handlers))

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

		handlerInstances[idx].Func = funcValue
		handlerInstances[idx].ParamTypes = paramTypes
	}

	return &handlerInstances, nil
}
