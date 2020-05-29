# gomvc
a simple golang mvc backend web service template, using gin-gonic/gin.   

与其说这是一个框架，不如是提供一个简单的golang http协议的后端服务模板  
相对于golang其他的框架，gomvc风格偏重go灵活的模块化，不做限制框架的约束， gomvc目标只是提供服务模板，不做复杂实现，方便在此上进一步开发。   
编译：执行make生成可执行文件和相关配置到output目录

## 特点
   
gin兼容：基于gin-gonic/gin上二次开发，完全支持gin使用方式。
```
路由添加：utils.AddRoute("GET", "/rest/example/add", &action.Api{}, action.AddExample)
```

简单实用：执行流base类Execute定义请求工作流。层次调用关系：action->service->model->dao  
```
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

	//step3:api逻辑回调
	if err := cb(a.ctx); err != nil {
		a.fail(err)
	} else {
		a.success()
	}
}

```
自动构造：请求和返回值，使用结构描述（类似pb），自动反射和校验请求返回值对象//持续完善中  
```
request:
type Add struct {
	Key    string `req:"required"`
	Value  string `req:"required"`
	Detail string `req:"optional"`
}

response:
type Add struct {
	AffectedNum int64 `req:"required"`
}
```

半自动的db orm封装：对orm的易用性和支持复杂sql的功能性平衡//持续完善中    
