package db

import (
	"fmt"

	"github.com/neil-peng/gomvc/utils"
)

type Field struct {
	fieldItems    []string
	fieldRawItems []string
	valueItems    []interface{}
}

func (f *Field) field(fi string) {
	f.fieldItems = append(f.fieldItems, fi)
}

func (f *Field) fieldRaw(fi string) {
	f.fieldRawItems = append(f.fieldRawItems, fi)
}

func (f *Field) fields(fis []string) {
	for _, fi := range fis {
		f.field(fi)
	}
}

func (f *Field) value(v interface{}) {
	f.valueItems = append(f.valueItems, v)

}

func (f *Field) values(vs ...interface{}) {
	for _, v := range vs {
		f.value(v)
	}
}

func (f *Field) fieldValue(k string, v interface{}) {
	f.field(k)
	f.value(v)
}

func (f *Field) fieldValues(k string, args ...interface{}) {
	if len(args)%2 == 0 {
		panic("args num is invlid")
	}
	var fN string = k
	var fV interface{}
	for i, arg := range args {
		if i%2 == 0 {
			fV = arg
			f.fieldValue(fN, fV)
		} else {
			fN, _ = arg.(string)
		}
	}
}

func (f *Field) formatFields() string {
	var partSql string
	for _, fieldItem := range f.fieldItems {
		if len(partSql) == 0 {
			partSql = fieldItem
		} else {
			partSql += fmt.Sprintf(", %s", fieldItem)
		}
	}
	return partSql
}

func (f *Field) formatValues() string {
	var partSql string
	for _, valueItem := range f.valueItems {
		if valueItemStr, ok := valueItem.(string); ok {
			valueItem = "'" + utils.Addslashes(valueItemStr) + "'"
		}
		if len(partSql) == 0 {
			partSql = fmt.Sprintf("%v", valueItem)
		} else {
			partSql += fmt.Sprintf(", %v", valueItem)
		}
	}
	return partSql
}

func (f *Field) formatFieldValues() string {
	var partSql string
	if len(f.fieldItems) != len(f.valueItems) {
		panic("fields num not equal values")
	}

	for index, fieldItem := range f.fieldItems {
		valueItem := f.valueItems[index]
		if !utils.InStringArray(f.fieldRawItems, fieldItem) {
			if valueItemStr, ok := valueItem.(string); ok {
				valueItem = "'" + utils.Addslashes(valueItemStr) + "'"
			}
		}
		if len(partSql) == 0 {
			partSql = fmt.Sprintf("%s=%v", fieldItem, valueItem)
		} else {
			partSql += fmt.Sprintf(", %s=%v", fieldItem, valueItem)
		}
	}
	return partSql
}

func (f *Field) FormatIntoMap() map[string]string {
	if len(f.fieldItems) != len(f.valueItems) {
		panic("fields num not equal values")
	}
	fieldMap := make(map[string]string, len(f.fieldItems))
	for index, fieldItem := range f.fieldItems {
		valueItem := f.valueItems[index]
		fieldMap[fieldItem] = fmt.Sprintf("%v", valueItem)
	}
	return fieldMap
}
