package action

import (
	"github.com/neil-peng/gomvc/model"
	"github.com/neil-peng/gomvc/request"
	"github.com/neil-peng/gomvc/response"
	"github.com/neil-peng/gomvc/utils"
)

func GetExample(ctx *utils.Context) error {
	//1. 参数解析
	req := &request.Get{}
	if err := (&request.Request{Context: ctx}).Valid(req); err != nil {
		ctx.Warn("invalid param req:%v, err:%v", req, err)
		return err
	}

	//2. 业务逻辑
	value, detail, err := (&model.Example{Context: ctx}).Get(req.Key)
	if err != nil {
		ctx.Warn("add %+v error, err:%v", req, err)
		return err
	}

	//3. 填返回值
	return (&response.Response{Context: ctx}).Format(&response.Get{
		Value:  value,
		Detail: detail,
	})
}
