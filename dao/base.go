package dao

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/utils"
)

type Base struct {
	*utils.Context
	TableView string
}

func (b *Base) GetTableView() string {
	return b.TableView
}

//cache json转table类型
func (b *Base) BuildImplicitField(m map[string]interface{}) (interface{}, error) {
	var daoIns interface{}
	switch b.TableView {
	case conf.TABLE_EXAMPLE:
		daoIns = &TableExample{}
	default:
		b.Critical("build invalid tableView from %s", b.TableView)
		return nil, errors.New(conf.ERROR_FIELD_SCHEME_INVALID)
	}

	valueRf := reflect.ValueOf(daoIns).Elem()
	typeRf := valueRf.Type()
	for i := 0; i < valueRf.NumField(); i++ {
		fieldValue := valueRf.Field(i)
		dbTag := typeRf.Field(i).Tag.Get("db")
		if srcValue, ok := m[dbTag]; ok {
			b.specInterface(srcValue, &fieldValue)
		}
	}
	return daoIns, nil
}

func (b *Base) BuildFields(ms []map[string]string) ([]interface{}, error) {
	var allItem []interface{}
	for _, m := range ms {
		oneItem, err := b.BuildField(m)
		if err != nil {
			return nil, err
		}
		allItem = append(allItem, oneItem)
	}
	return allItem, nil
}

//将srcvalue模糊类型自动转成确定类型的方法
func (b *Base) specInterface(srcValue interface{}, dstFieldValue *reflect.Value) {
	srcType := reflect.TypeOf(srcValue).Kind()
	dstType := dstFieldValue.Type().Kind()
	switch realSrcValue := srcValue.(type) {
	case string:
		switch dstType {
		case reflect.Int:
			dbIntValue, _ := strconv.ParseInt(realSrcValue, 10, 0)
			dstFieldValue.SetInt(dbIntValue)
		case reflect.Int64:
			dbInt64Value, _ := strconv.ParseInt(realSrcValue, 10, 64)
			dstFieldValue.SetInt(dbInt64Value)
		case reflect.Uint:
			dbUintValue, _ := strconv.ParseUint(realSrcValue, 10, 0)
			dstFieldValue.SetUint(dbUintValue)
		case reflect.Uint64:
			dbUint64Value, _ := strconv.ParseUint(realSrcValue, 10, 64)
			dstFieldValue.SetUint(dbUint64Value)
		case reflect.Float32, reflect.Float64:
			dbFloatValue, _ := strconv.ParseInt(realSrcValue, 10, 64)
			dstFieldValue.SetFloat(float64(dbFloatValue))
		case reflect.String:
			dstFieldValue.SetString(realSrcValue)
		default:
			b.Critical("[reflect field failed] [srctype:%s] [dsttype:%s]", srcType, dstType)
			panic(errors.New(conf.ERROR_FIELD_SCHEME_INVALID))
		}
	case float32, float64:
		switch dstType {
		case reflect.Int, reflect.Int64:
			dbIntValue := int64(reflect.ValueOf(realSrcValue).Float())
			dstFieldValue.SetInt(dbIntValue)
		case reflect.Uint, reflect.Uint64:
			dbUintValue := uint64(reflect.ValueOf(realSrcValue).Float())
			dstFieldValue.SetUint(dbUintValue)
		case reflect.String:
			dstFieldValue.SetString(fmt.Sprintf("%v", reflect.ValueOf(realSrcValue).Float()))
		default:
			b.Critical("[reflect field failed] [srctype:%s] [dsttype:%s]", srcType, dstType)
			panic(errors.New(conf.ERROR_FIELD_SCHEME_INVALID))
		}
	case int, int8, int16, int32, int64:
		switch dstType {
		case reflect.Int, reflect.Int64:
			dbIntValue := int64(reflect.ValueOf(realSrcValue).Int())
			dstFieldValue.SetInt(dbIntValue)
		case reflect.Uint, reflect.Uint64:
			dbUintValue := uint64(reflect.ValueOf(realSrcValue).Int())
			dstFieldValue.SetUint(dbUintValue)
		case reflect.String:
			dstFieldValue.SetString(fmt.Sprintf("%v", reflect.ValueOf(realSrcValue).Int()))
		default:
			b.Critical("[reflect field failed] [srctype:%s] [dsttype:%s]", srcType, dstType)
			panic(errors.New(conf.ERROR_FIELD_SCHEME_INVALID))
		}
	case uint8, uint16, uint32, uint64:
		switch dstType {
		case reflect.Int, reflect.Int64:
			dbIntValue := int64(reflect.ValueOf(realSrcValue).Uint())
			dstFieldValue.SetInt(dbIntValue)
		case reflect.Uint, reflect.Uint64:
			dbUintValue := uint64(reflect.ValueOf(realSrcValue).Uint())
			dstFieldValue.SetUint(dbUintValue)
		case reflect.String:
			dstFieldValue.SetString(fmt.Sprintf("%v", reflect.ValueOf(realSrcValue).Uint()))
		default:
			b.Critical("[reflect field failed] [srctype:%s] [dsttype:%s]", srcType, dstType)
			panic(errors.New(conf.ERROR_FIELD_SCHEME_INVALID))
		}
	default:
		b.Critical("[reflect field failed] [srctype:%s] [dsttype:%s]", srcType, dstType)
		panic(errors.New(conf.ERROR_FIELD_SCHEME_INVALID))
	}
}

func (b *Base) BuildField(m map[string]string) (interface{}, error) {
	var daoIns interface{}
	switch b.TableView {
	case conf.TABLE_EXAMPLE:
		daoIns = &TableExample{}
	default:
		b.Critical("build invalid tableView from %s", b.TableView)
		return nil, errors.New(conf.ERROR_FIELD_SCHEME_INVALID)
	}

	valueRf := reflect.ValueOf(daoIns).Elem()
	typeRf := valueRf.Type()
	for i := 0; i < valueRf.NumField(); i++ {
		fieldValue := valueRf.Field(i)
		dbTag := typeRf.Field(i).Tag.Get("db")
		if dbStrValue, ok := m[dbTag]; ok {
			b.specInterface(dbStrValue, &fieldValue)
		}
	}
	return daoIns, nil
}
