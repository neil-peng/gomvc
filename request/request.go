package request

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/utils"
)

type Request struct {
	*utils.Context
}

func (r *Request) Valid(req interface{}) error {
	reqElem := reflect.ValueOf(req).Elem()
	reqType := reqElem.Type()
	for i := 0; i < reqElem.NumField(); i++ {
		if len(reqType.Field(i).Tag.Get("req")) == 0 {
			continue
		}

		keyName := strings.ToLower(reqType.Field(i).Name)
		requiredKeyValue := utils.Trim(r.Query(keyName))
		if len(requiredKeyValue) == 0 {
			if reqType.Field(i).Tag.Get("req") == "required" {
				return fmt.Errorf("missing or emtpy required param %s", keyName)
			} else {
				continue
			}
		}
		r.PushNotice(keyName, requiredKeyValue)
		fv := reqElem.Field(i)
		switch fv.Type().Kind() {
		case reflect.Int:
			intValue, err := strconv.ParseInt(requiredKeyValue, 10, 0)
			if err != nil {
				return err
			}
			fv.SetInt(intValue)
		case reflect.Int64:
			int64Value, err := strconv.ParseInt(requiredKeyValue, 10, 64)
			if err != nil {
				return err
			}
			fv.SetInt(int64Value)
		case reflect.Uint:
			uintValue, err := strconv.ParseUint(requiredKeyValue, 10, 0)
			if err != nil {
				return err
			}
			fv.SetUint(uintValue)
		case reflect.Uint64:
			uint64Value, err := strconv.ParseUint(requiredKeyValue, 10, 64)
			if err != nil {
				return err
			}
			fv.SetUint(uint64Value)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseInt(requiredKeyValue, 10, 64)
			if err != nil {
				return err
			}
			fv.SetFloat(float64(floatValue))
		case reflect.String:
			fv.SetString(requiredKeyValue)
		default:
			panic(fmt.Errorf("%v", conf.ERROR_PARAM_ERROR))
		}
	}
	return nil
}
