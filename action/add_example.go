package action

import (
	"github.com/neil-peng/gomvc/model"
	"github.com/neil-peng/gomvc/request"
	"github.com/neil-peng/gomvc/response"
	"github.com/neil-peng/gomvc/utils"
)

func AddExample(ctx *utils.Context) error {
	//1. 参数解析
	req := &request.Add{}
	if err := (&request.Request{Context: ctx}).Valid(req); err != nil {
		ctx.Warn("invalid param req:%v, err:%v", req, err)
		return err
	}

	//2. 业务逻辑
	affectedNum, err := (&model.Example{Context: ctx}).Add(req.Key, req.Value, req.Detail)
	if err != nil {
		ctx.Warn("add %+v error, err:%v", req, err)
		return err
	}

	//3. 填返回值
	return (&response.Response{Context: ctx}).Format(&response.Add{
		AffectedNum: affectedNum,
	})
}
