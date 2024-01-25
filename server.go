package fastapi

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	mux *httprouter.Router
}

func New() *Server {
	return &Server{
		mux: httprouter.New(),
	}
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) POST(path string, handlers ...interface{}) error {

	handlerInstances, err := s.getHandlerInstances(handlers...)
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
