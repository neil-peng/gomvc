package utils

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-collections/collections/stack"
	"github.com/neil-peng/gomvc/conf"
)

type Context struct {
	*gin.Context
	*Logger

	nameServer    NameService
	callers       *stack.Stack
	costOpenClose bool
	sync.RWMutex
}

func (c *Context) SetNameService(servicer NameService) {
	c.nameServer = servicer
}

func (c *Context) GetServer(service string) (ip string, port string, err error) {
	return c.nameServer.GetServer(service)
}

func (c *Context) SetResponseBody(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if bodyMap := c.GetStringMap(conf.RESULT_BODY_MAP); bodyMap == nil {
		c.Set(conf.RESULT_BODY_MAP, map[string]interface{}{
			key: value,
		})
	} else {
		bodyMap[key] = value
	}
}

func (c *Context) GetResponseBody(key string) interface{} {
	c.RLock()
	defer c.RUnlock()
	if bodyMap := c.GetStringMap(conf.RESULT_BODY_MAP); bodyMap == nil {
		return bodyMap[key]
	}
	return nil
}

func (c *Context) GetJsonResponseBody() []byte {
	c.RLock()
	defer c.RUnlock()
	if bodyMap := c.GetStringMap(conf.RESULT_BODY_MAP); bodyMap != nil {
		res, err := JSONEncode(bodyMap)
		if err != nil {
			c.Warn("json response body fail, err:%v", err)
		}
		return res
	}
	c.Info("json response empty body")
	return []byte{}
}

func (c *Context) SetResponseHeader(key string, value string) {
	c.Lock()
	defer c.Unlock()
	if bodyMap := c.GetStringMapString(conf.RESULT_HEADER_MAP); bodyMap == nil {
		c.Set(conf.RESULT_HEADER_MAP, map[string]string{
			key: value,
		})
	} else {
		bodyMap[key] = value
	}
}

func (c *Context) GetResponseHeader(key string) string {
	c.RLock()
	defer c.RUnlock()
	if headerMap := c.GetStringMapString(conf.RESULT_HEADER_MAP); headerMap != nil {
		return headerMap[key]
	}
	return ""
}

func (c *Context) ListResponseHeader() map[string]string {
	c.RLock()
	defer c.RUnlock()
	return c.GetStringMapString(conf.RESULT_HEADER_MAP)
}

func (c *Context) LogId() string {
	if logId := c.Query("logid"); len(logId) > 0 {
		return logId
	}

	if logId := c.GetString("logid"); len(logId) > 0 {
		return logId
	}

	c.Set("logid", RandId())
	return c.GetString("logid")
}

func (c *Context) CloseCostGather() {
	c.Lock()
	defer c.Unlock()
	c.costOpenClose = true
	return
}

func (c *Context) OpenCostGather() {
	c.Lock()
	defer c.Unlock()
	c.costOpenClose = false
	return
}

func (c *Context) IfCloseGather() bool {
	c.RLock()
	defer c.RUnlock()
	return c.costOpenClose
}

func (c *Context) StatusStart() {
	if c.IfCloseGather() {
		return
	}

	if c.callers == nil {
		c.callers = stack.New()
	}

	startTime := time.Now()
	pc, _, _, ok := runtime.Caller(1)
	callFunc := runtime.FuncForPC(pc)
	if ok && callFunc != nil {
		var funcName string
		fullPathArr := strings.Split(callFunc.Name(), ".")
		if l := len(fullPathArr); l > 0 {
			funcName = fullPathArr[l-1]
		} else {
			funcName = callFunc.Name()
		}
		c.callers.Push(map[string]time.Time{
			funcName: startTime,
		})
	}
}

func (c *Context) StatusEnd() int {
	if c.IfCloseGather() {
		return 0
	}
	if c.callers.Len() == 0 {
		c.Warn("[status format error]")
		return 0
	}

	callFuncInfo := c.callers.Pop().(map[string]time.Time)
	for funcName, startTime := range callFuncInfo {
		cost := time.Since(startTime) / time.Millisecond
		c.PushNotice(funcName, fmt.Sprintf("%dms", cost))
		return int(cost)
	}
	return 0
}
