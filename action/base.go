package action

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/utils"
)

type Api struct {
	ctx *utils.Context
}

//api处理流程
func (a *Api) Execute(ctx *gin.Context, cb utils.ApiCb) {
	defer a.finish()

	//step1:创建请求上下文
	if err := a.newContext(ctx); err != nil {
		a.fail(err)
		return
	}

	//step2:api初始化工作
	if err := a.init(); err != nil {
		a.fail(err)
		return
	}

	//step3:api逻辑回调（可能会抛出panic，通过error处理结果，未定义异常会出panic）
	if err := cb(a.ctx); err != nil {
		a.fail(err)
	} else {
		a.success()
	}
}

func (a *Api) newContext(ctx *gin.Context) error {
	a.ctx = &utils.Context{
		Context: ctx,
		Logger:  utils.NewLogger(),
	}

	clientIp := a.ctx.ClientIP()
	hostIp, _ := utils.GetHostIp()

	a.ctx.SetNameService(&utils.IpServer)
	a.ctx.Set(conf.CLIENT_IP, clientIp)
	a.ctx.Set(conf.SERVER_IP, hostIp)
	a.ctx.Set(conf.API_TIME, time.Now())

	a.ctx.SetBaseInfo("logid", a.ctx.LogId())
	a.ctx.PushNotice("client", clientIp)
	return nil
}

func (a *Api) init() error {

	return nil
}

func (a *Api) success() {
	a.ctx.SetResponseBody(conf.API_STATUS, errors.New(conf.NO_ERROR))
	a.ctx.SetResponseBody(conf.LOG_ID, a.ctx.LogId())
	a.ctx.SetResponseBody(conf.ERR_CODE, "0")
	a.ctx.SetResponseBody(conf.ERR_MSG, "success")
	a.ctx.Status(200)

	a.ctx.PushNotice(conf.HTTP_CODE, "200")
	a.ctx.PushNotice(conf.ERR_MSG, "success")
	a.ctx.PushNotice(conf.ERR_CODE, "0")
}

func (a *Api) fail(err error) {
	errMsg := conf.GetHttpMsg(err.Error())
	a.ctx.Set("status", conf.GetHttpCode(err.Error()))
	a.ctx.SetResponseBody(conf.ERR_MSG, errMsg)
	a.ctx.SetResponseBody(conf.ERR_CODE, err.Error())
	a.ctx.SetResponseBody("log_id", a.ctx.LogId())

	a.ctx.PushNotice(conf.HTTP_CODE, conf.GetHttpCode(err.Error()))
	a.ctx.PushNotice(conf.ERR_MSG, errMsg)
	a.ctx.PushNotice(conf.ERR_CODE, err.Error())
}

func (a *Api) finish() {
	if r := recover(); r != nil {
		//非预期的异常全部转化成特性错误错误，501
		a.ctx.Critical("panic err:%v, stacktrace:%s", r, string(debug.Stack()))
		a.fail(errors.New(conf.ERROR_NETWORK_ERROR))
	}

	//填写返回的header
	for k, v := range a.ctx.ListResponseHeader() {
		a.ctx.Header(k, v)
	}
	apiResult := a.ctx.GetJsonResponseBody()
	a.ctx.Info("api return:%s", apiResult)
	a.ctx.Data(a.ctx.GetInt("status"), "text/html;charset=utf8", apiResult)
	apiTime := a.ctx.MustGet(conf.API_TIME).(time.Time)
	apiCost := time.Since(apiTime)
	cost := fmt.Sprintf("%d", apiCost/1000/1000)
	a.ctx.PushNotice("referer", a.ctx.GetHeader("Referer"))
	a.ctx.Notice("[process_time:%sms]", cost)
}

func (a *Api) New() utils.ApiActor {
	return &Api{}
}
