package fastapi

import (
	"encoding/json"
	"mime/multipart"
	"reflect"
)

type Handler struct {
	handlerFunc reflect.Value
	paramType   reflect.Type
}

func (h Handler) createParam(ctx *Context, pType reflect.Type, pVal reflect.Value) error {
	for i := 0; i < pVal.NumField(); i++ {
		if pType.Field(i).Tag.Get("body") == "json" {
			decoder := json.NewDecoder(ctx.Request.Body)
			b := pVal.Field(i).Addr().Interface()
			err := decoder.Decode(b)
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("body") == "urlencoded" {
			for j := 0; j < pVal.Field(i).NumField(); j++ {
				err := setValue(pVal.Field(i).Field(j), ctx.Request.FormValue(pType.Field(i).Type.Field(j).Tag.Get("form")))
				if err != nil {
					return err
				}
			}

		} else if pType.Field(i).Tag.Get("body") == "multipart" {
			for j := 0; j < pVal.Field(i).NumField(); j++ {
				fileType := reflect.TypeOf((*multipart.File)(nil)).Elem()

				if pVal.Field(i).Field(j).Type().ConvertibleTo(fileType) {
					file, _, err := ctx.Request.FormFile(pType.Field(i).Type.Field(j).Tag.Get("form"))
					if err != nil {
						return err
					}
					pVal.Field(i).Field(j).Set(reflect.ValueOf(file))

				} else {
					err := setValue(pVal.Field(i).Field(j), ctx.Request.FormValue(pType.Field(i).Type.Field(j).Tag.Get("form")))
					if err != nil {
						return err
					}
				}
			}

		} else if pType.Field(i).Type.Kind() == reflect.Struct {
			err := h.createParam(ctx, pType.Field(i).Type, pVal.Field(i))
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("path") != "" {
			err := setValue(pVal.Field(i), ctx.params.ByName(pType.Field(i).Tag.Get("path")))
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("query") != "" {
			err := setValue(pVal.Field(i), ctx.Request.URL.Query().Get(pType.Field(i).Tag.Get("query")))
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("header") != "" {
			err := setValue(pVal.Field(i), ctx.Request.Header.Get(pType.Field(i).Tag.Get("header")))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h Handler) handle(ctx *Context) error {
	param := reflect.New(h.paramType).Elem()

	err := h.createParam(ctx, h.paramType, param)
	if err != nil {
		return err
	}

	result := h.handlerFunc.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		param,
	})

	res := result[0].Interface()
	if res != nil {
		return res.(error)
	}

	return nil
}
