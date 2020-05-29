package utils

import (
	"github.com/gin-gonic/gin"
)

type ApiCb func(context *Context) error

type ApiActor interface {
	New() ApiActor
	Execute(*gin.Context, ApiCb)
}

var g = gin.Default()

func AddRoute(method string, path string, apiAct ApiActor, cb ApiCb) {
	g.Handle(method, path, func(c *gin.Context) {
		apiAct.New().Execute(c, cb)
	})
}

func RunServer(server string) {
	Notice("run:%v", g.Run(server))
}
