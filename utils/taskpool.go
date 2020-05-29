package utils

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrPoolTimeOut = errors.New("add to pool timeout")

//任务回调函数
type FuncCb func(param interface{}) (cbResult interface{}, cbErr error)

//任务结果处理函数，可以针对有序消息的进行结果汇总（有序消息需要保证单进多出方式）
type FuncOut func(msgId int64, cbResult interface{}, cbErr error) (err error)

type QueueItem struct {
	id    int64
	param interface{}
}

type TaskPool struct {
	Ctx           *Context
	Size          int     //队列大小，db类查询建议值10
	PoolTimeOutMs int     //投递队列超时时间
	Cb            FuncCb  //任务回调
	Out           FuncOut //汇总任务返回，可选
	LooseCheck    bool    //单次任务失败是否认为总任务失败，可选，默认严格校验

	id    int64
	queue chan *QueueItem
	err   error
	stop  bool
	sync.WaitGroup
	sync.RWMutex
}

func (t *TaskPool) Init() *TaskPool {
	if t.Ctx != nil {
		t.Ctx.CloseCostGather()
	}
	t.queue = make(chan *QueueItem, t.Size)
	for i := 0; i < t.Size; i++ {
		t.Add(1)
		go func() {
			defer t.Done()
			for queueMsg := range t.queue {
				if t.IfStop() {
					break
				}
				subResult, err := t.Cb(queueMsg.param)
				if t.Out == nil {
					if t.LooseCheck {
						continue
					}
					if !t.LooseCheck && err != nil {
						t.SetStop(err)
						break
					}
				} else {
					t.Lock()
					outErr := t.Out(queueMsg.id, subResult, err)
					t.Unlock()
					if !t.LooseCheck && outErr != nil {
						t.SetStop(outErr)
						break
					}
				}
			}
		}()
	}
	return t
}

func (t *TaskPool) IfStop() bool {
	t.RLock()
	defer t.RUnlock()
	return t.stop
}

func (t *TaskPool) SetStop(err error) {
	t.Lock()
	defer t.Unlock()
	t.stop = true
	t.err = err
	return
}

func (t *TaskPool) Process(param interface{}) error {
	id := atomic.LoadInt64(&t.id)
	defer atomic.AddInt64(&t.id, 1)

	if t.PoolTimeOutMs <= 0 {
		t.queue <- &QueueItem{id, param}
		return nil
	}

	select {
	case t.queue <- &QueueItem{id, param}:
	case <-time.After(time.Duration(t.PoolTimeOutMs) * time.Millisecond):
		t.err = ErrPoolTimeOut
		return t.err
	}
	return nil
}

func (t *TaskPool) Join() error {
	close(t.queue)
	t.Wait()
	if t.Ctx != nil {
		t.Ctx.OpenCostGather()
	}
	return t.err
}
