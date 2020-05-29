package response

import (
	"reflect"
	"strings"

	"github.com/neil-peng/gomvc/utils"
)

type Response struct {
	*utils.Context
}

func (r *Response) Format(res interface{}) error {
	if res == nil {
		return nil
	}
	reqElem := reflect.ValueOf(res).Elem()
	reqType := reqElem.Type()
	for i := 0; i < reqElem.NumField(); i++ {
		fieldName := strings.ToLower(reqType.Field(i).Name)
		r.SetResponseBody(fieldName, reqElem.Field(i).Interface())
	}
	return nil
}
