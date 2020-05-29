# gomvc
a simple golang mvc backend web service template, using gin-gonic/gin.   

与其说这是一个框架，不如是提供一个简单的golang http协议的后端服务模板  
相对于golang其他的框架，gomvc风格偏重go灵活的模块化，不做限制框架的约束， gomvc目标只是提供服务模板，不做复杂实现，方便在此上进一步开发。  
  
简单实用：执行流 action->service->model->dao  
自动构造：请求和返回值，使用结构描述（类似pb），自动反射和校验请求返回值对象//持续完善中  
半自动的db orm封装：对orm的易用性和支持复杂sql的功能性平衡//持续完善中 
gin兼容：框架基于gin-gonic/gin上的二次开发
