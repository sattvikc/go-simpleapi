package fastapi

import (
	"fmt"
	"reflect"
	"strconv"
)

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
