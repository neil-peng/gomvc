package dao

import (
	//"errors"
	//"fmt"

	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/lib/db"
	"github.com/neil-peng/gomvc/utils"
)

//表描述，用于将map[string]interface{}反射具体实例
type TableExample struct {
	//!golang tag之间务必里空格分割，不识别tab,https://golang.org/pkg/reflect/#StructTag
	Key    string `db:"id"`
	Value  string `db:"value"`
	Detail string `db:"detail"`
}

//描述表视图，关联到表描述
type TableExampleView struct {
	Base
	Err error
}

func NewTableExampleView(ctx *utils.Context) *TableExampleView {
	return &TableExampleView{
		Base: Base{TableView: conf.TABLE_EXAMPLE, Context: ctx},
	}
}

func (t *TableExampleView) Add(key, value, detail string) (int64, error) {
	//函数计时
	t.StatusStart()
	defer t.StatusEnd()
	return db.New(t).FieldValues("id", key, "value", value, "detail", detail).Insert()
}

func (t *TableExampleView) Get(key string) (*TableExample, error) {
	//函数计时
	t.StatusStart()
	defer t.StatusEnd()
	tableExamples, err := db.New(t).Field("*").SetCond("id=\"%s\"", key).SelectToBuild()
	if err != nil {
		return nil, err
	}
	if len(tableExamples) > 0 {
		return tableExamples[0].(*TableExample), nil
	}
	return nil, nil
}
