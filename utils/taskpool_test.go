package utils

import (
	"errors"
	"fmt"
	"gotest.tools/assert"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestOnlyCbTaskPool(t *testing.T) {
	var err error
	var mu sync.Mutex
	var result []int
	p := (&TaskPool{Size: 10, PoolTimeOutMs: 2000,
		Cb: func(param interface{}) (interface{}, error) {
			time.Sleep(time.Duration(Rand(0, 1000)) * time.Millisecond)
			index := param.(int)
			subResult := 2 * index
			mu.Lock()
			result = append(result, subResult)
			mu.Unlock()
			return subResult, nil
		}}).Init()

	for i := 0; i < 10; i++ {
		if err = p.Process(i); err != nil {
			fmt.Printf("[process error] [id:%d] [err:%+v]\n", i, err)
			break
		}
	}

	p.Join()
	fmt.Printf("result:%+v\n", result)
	assert.NilError(t, err)
	sort.Ints(result)
	assert.Equal(t, true, reflect.DeepEqual(result, []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18}))
	return
}

func TestCbOutTaskPool(t *testing.T) {
	var err error
	result := make([]int, 10)
	p := (&TaskPool{
		Size:          10,
		PoolTimeOutMs: 2000,
		Cb: func(param interface{}) (interface{}, error) {
			time.Sleep(time.Duration(Rand(0, 1000)) * time.Millisecond)
			index := param.(int)
			subResult := index
			return subResult, nil
		},
		Out: func(msgId int64, cbResult interface{}, err error) error {
			if err != nil {
				return err
			}
			result[msgId] = cbResult.(int)
			return nil
		},
	}).Init()

	for i := 0; i < 10; i++ {
		if subErr := p.Process(i); subErr != nil {
			fmt.Printf("[process error] [id:%d] [err:%+v]\n", i, subErr)
			break
		}
	}

	err = p.Join()
	fmt.Printf("result:%+v\n", result)
	assert.NilError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(result, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
	return
}

func TestExceptionCbOutTaskPool(t *testing.T) {
	var err error
	result := make([]int, 10)
	p := (&TaskPool{
		Size:          10,
		PoolTimeOutMs: 2000,
		Cb: func(param interface{}) (interface{}, error) {
			index := param.(int)
			if index == 5 {
				return nil, errors.New("mannual error!")
			}
			subResult := index
			return subResult, nil
		},
		Out: func(msgId int64, cbResult interface{}, err error) error {
			if err != nil || cbResult == nil {
				return err
			}
			result[msgId] = cbResult.(int)
			return nil
		},
	}).Init()

	for i := 0; i < 10; i++ {
		if subErr := p.Process(i); subErr != nil {
			fmt.Printf("[process error] [id:%d] [err:%+v]\n", i, subErr)
			break
		}
	}

	err = p.Join()
	fmt.Printf("result:%+v\n", result)
	assert.Equal(t, true, err != nil)
	return
}

func TestTimeOutCbOutTaskPool(t *testing.T) {
	var err error
	result := make([]int, 12)
	p := (&TaskPool{
		Size:          2,
		PoolTimeOutMs: 200,
		Cb: func(param interface{}) (interface{}, error) {
			time.Sleep(2 * time.Second)
			index := param.(int)
			subResult := index
			return subResult, nil
		},
		Out: func(msgId int64, cbResult interface{}, err error) error {
			if err != nil || cbResult == nil {
				return err
			}
			result[msgId] = cbResult.(int)
			return nil
		},
	}).Init()

	for i := 0; i < 12; i++ {
		if subErr := p.Process(i); subErr != nil {
			fmt.Printf("[process error] [id:%d] [err:%+v]\n", i, subErr)
			break
		} else {
			fmt.Printf("[process succ] [id:%d]\n", i)
		}
	}

	err = p.Join()
	fmt.Printf("result:%+v\n", result)
	assert.Equal(t, true, err != nil)
	return
}
