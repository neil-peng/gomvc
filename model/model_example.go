package model

import (
	"github.com/neil-peng/gomvc/dao"
	"github.com/neil-peng/gomvc/utils"
)

type Example struct {
	*utils.Context
}

func (e *Example) Add(key, value, detail string) (int64, error) {
	return dao.NewTableExampleView(e.Context).Add(key, value, detail)
}

func (e *Example) Get(key string) (string, string, error) {
	if tableExample, err := dao.NewTableExampleView(e.Context).Get(key); err != nil {
		return "", "", err
	} else {
		return tableExample.Value, tableExample.Detail, nil
	}
}
