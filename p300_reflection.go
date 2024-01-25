package fastapi

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
)

func populateValueFromTypeUsingContext(ctx *Context, pType reflect.Type, pVal reflect.Value) error {
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
			err := populateValueFromTypeUsingContext(ctx, pType.Field(i).Type, pVal.Field(i))
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("path") != "" {
			err := setValue(pVal.Field(i), ctx.params.ByName(pType.Field(i).Tag.Get("path")))
			if err != nil {
				return err
			}

		} else if pType.Field(i).Tag.Get("query") != "" {
			if pVal.Field(i).Type().Kind() != reflect.Ptr || ctx.Request.URL.Query().Get(pType.Field(i).Tag.Get("query")) != "" {
				err := setValue(pVal.Field(i), ctx.Request.URL.Query().Get(pType.Field(i).Tag.Get("query")))
				if err != nil {
					return err
				}
			}

		} else if pType.Field(i).Tag.Get("header") != "" {
			if pVal.Field(i).Type().Kind() != reflect.Ptr || ctx.Request.Header.Get(pType.Field(i).Tag.Get("header")) != "" {
				err := setValue(pVal.Field(i), ctx.Request.Header.Get(pType.Field(i).Tag.Get("header")))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func setValue(valueObj reflect.Value, value string) error {

	if valueObj.Kind() == reflect.Ptr {
		valueObj.Set(reflect.New(valueObj.Type().Elem()))
		valueObj = valueObj.Elem()
	}

	switch valueObj.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Parse value as integer
		intValue, err := strconv.ParseInt(value, 10, valueObj.Type().Bits())
		if err != nil {
			return err
		}
		valueObj.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Parse value as unsigned integer
		uintValue, err := strconv.ParseUint(value, 10, valueObj.Type().Bits())
		if err != nil {
			return err
		}
		valueObj.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		// Parse value as float
		floatValue, err := strconv.ParseFloat(value, valueObj.Type().Bits())
		if err != nil {
			return err
		}
		valueObj.SetFloat(floatValue)
	case reflect.Bool:
		// Parse value as boolean
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		valueObj.SetBool(boolValue)
	case reflect.String:
		// Set value as string
		valueObj.SetString(value)
	default:
		return fmt.Errorf("unsupported type: %v", valueObj.Kind())
	}

	return nil
}
