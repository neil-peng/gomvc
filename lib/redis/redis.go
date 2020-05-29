package redis

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/utils"
)

type RedisPool struct {
	_pool *redis.Pool
	_addr string
}

type Redis struct {
	*utils.Context
	_name  string
	_redis *RedisPool
}

type NamedRedisPool struct {
	_redisPoolMap map[string]*RedisPool
	sync.Mutex
}

var namedRedisPool *NamedRedisPool

func init() {
	namedRedisPool = &NamedRedisPool{
		_redisPoolMap: map[string]*RedisPool{},
	}
}

func (n *NamedRedisPool) addPool(name string, pool *RedisPool) {
	n.Lock()
	defer n.Unlock()
	n._redisPoolMap[name] = pool
}

func (n *NamedRedisPool) getPool(name string) *RedisPool {
	n.Lock()
	defer n.Unlock()
	return n._redisPoolMap[name]
}

func (n *NamedRedisPool) removePool(name string) {
	n.Lock()
	defer n.Unlock()
	delete(n._redisPoolMap, name)
}

func (r *Redis) Name(redisServiceName string) *Redis {
	r._name = redisServiceName
	var err error
	if err = r.createNamePool(redisServiceName, false); err != nil {
		r.Critical("get redis fail, service:%s, err:%v", redisServiceName, err)
		return nil
	}
	return r
}

func (r *Redis) createNamePool(redisServiceName string, isFresh bool) error {
	var redisPool *RedisPool
	if rp := namedRedisPool.getPool(redisServiceName); rp != nil {
		if isFresh == false {
			r._redis = rp
			return nil
		}
		if r._redis == rp {
			rp._pool.Close()
			namedRedisPool.removePool(redisServiceName)
			r.Info("close redis pool, addr:%s", rp._addr)
		} else {
			r.Info("ignore competition redis pool, reuse addr:%s", rp._addr)
			redisPool = rp
			return nil
		}
	}
	defaultIp, defaultPort, err := r.GetServer(redisServiceName)
	if err != nil {
		r.Critical("init redis pool failed, service:%s, err:%v", redisServiceName, err)
		return err
	}
	pool := &redis.Pool{
		MaxIdle:     conf.REDIS_POOL_INT_MAX_IDLE_NUMS,
		MaxActive:   conf.REDIS_POOL_INT_MAX_ACTIVE_NUMS,
		IdleTimeout: time.Duration(conf.REDIS_POOL_INT_IDLE_TIMEOUT) * time.Second,
		Dial: func() (redis.Conn, error) {
			ip, port, err := r.GetServer(redisServiceName)
			if err != nil {
				r.Critical("init redis pool failed, service:%s, err:%v", redisServiceName, err)
				ip = defaultIp
				port = defaultPort
			}
			address := ip + ":" + port
			c, err := redis.DialTimeout("tcp", address,
				conf.REDIS_CONNECT_TIMEOUTMS*time.Millisecond,
				conf.REDIS_READ_TIMEOUTMS*time.Millisecond,
				conf.REDIS_WRITE_TIMEOUTMS*time.Millisecond)
			if err != nil {
				r.Warn("redis dail %s failed", address)
				return nil, err
			}
			return c, err
		},
		Wait: conf.REDIS_POOL_BOOL_WAIT,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	redisPool = &RedisPool{}
	redisPool._pool = pool

	if rp := namedRedisPool.getPool(redisServiceName); rp != nil {
		redisPool = rp
	} else {
		namedRedisPool.addPool(redisServiceName, redisPool)
	}

	r.Info("init redis pool succ, service:%s", redisServiceName)
	return nil
}

func (r *Redis) get(key string) (string, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("get connection failed, active nums:%d, error:%s",
			r._redis._pool.ActiveCount(), err)
		r.createNamePool(r._name, true)
		return "", errors.New(conf.ERROR_CONN_CACHE)
	}

	res, err := conn.Do("GET", key)
	if err != nil {
		r.Warn("[do get failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return "", errors.New(conf.ERROR_GET_CACHE)
	}

	v, err := redis.String(res, err)
	if err == nil {
		r.Debug("[get key %s successed] [value:%s] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	if err == redis.ErrNil {
		r.Debug("[get key %s empty] [active nums:%d]",
			key, r._redis._pool.ActiveCount())
		return "", nil
	}
	return "", errors.New(conf.ERROR_GET_CACHE)
}

func (r *Redis) Get(key string) (string, error) {
	if r._redis == nil {
		return "", errors.New(conf.ERROR_CONN_CACHE)
	}

	var res string
	var err error
	for i := 0; i < conf.RETRY; i++ {
		res, err = r.get(key)
		if err == nil {
			return res, nil
		}
	}
	r.Warn("[get key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return res, err
}

func (r *Redis) set(key string, value string) error {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("SET", key, value)
	if err != nil {
		r.Warn("[do set failed, reconnect [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.String(res, err)
	if err == nil && strings.EqualFold(v, "OK") {
		r.Debug("[set key %s successed] [value:%s] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return nil
	}
	return errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Set(key string, value string) error {
	if r._redis == nil {
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	for i := 0; i < conf.RETRY; i++ {
		err = r.set(key, value)
		if err == nil {
			return nil
		}
	}
	r.Warn("[set key:%s failed] [value:%s] [error:%s] [active nums:%d]",
		key, value, err, r._redis._pool.ActiveCount())
	return err
}

func (r *Redis) setEx(key string, expire int, value string) error {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("SETEX", key, expire, value)
	if err != nil {
		r.Warn("[do setex failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.String(res, err)
	if err == nil && strings.EqualFold(v, "OK") {
		r.Debug("[setex key %s successed] [expire:%d] [value:%s] [active nums:%d]",
			key, expire, v, r._redis._pool.ActiveCount())
		return nil
	}

	r.Info("[setex key:%s failed] [value:%s] [error:%s] [active nums:%d]",
		key, value, err, r._redis._pool.ActiveCount())
	return errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) SetEx(key string, expire int, value string) error {
	if r._redis == nil {
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	for i := 0; i < conf.RETRY; i++ {
		err = r.setEx(key, expire, value)
		if err == nil {
			return nil
		}
	}
	r.Warn("[setex key:%s failed] [expire:%d] [value:%s] [error:%s] [active nums:%d]",
		key, expire, value, err, r._redis._pool.ActiveCount())
	return err
}

func (r *Redis) setNx(key string, value ...interface{}) (error, int) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_CONN_CACHE), -1
	}

	res, err := conn.Do("SET", append([]interface{}{key}, value...)...)
	if err != nil {
		r.Warn("[do setnx failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_SET_CACHE), -1
	}

	if res == nil && err == nil {
		return nil, 0
	}
	v, err := redis.String(res, err)
	if err == nil && strings.EqualFold(v, "OK") {
		r.Debug("[setnx key %s successed] [value:%s] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return nil, 1
	}

	r.Info("[setnx key:%s failed] [value:%s] [error:%s] [active nums:%d]",
		key, value, err, r._redis._pool.ActiveCount())
	return errors.New(conf.ERROR_SET_CACHE), -1
}

func (r *Redis) SetNx(key string, value ...interface{}) (error, int) {
	if r._redis == nil {
		return errors.New(conf.ERROR_CONN_CACHE), -1
	}
	var err error
	for i := 0; i < conf.RETRY; i++ {
		err, res := r.setNx(key, value...)
		if err == nil {
			return nil, res
		}
	}
	r.Warn("[setnx key:%s failed] [value:%s] [error:%s] [active nums:%d]",
		key, value, err, r._redis._pool.ActiveCount())
	return err, -1
}

func (r *Redis) expire(key string, timeout int) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("EXPIRE", key, timeout)
	if err != nil {
		r.Warn("[do expire failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)

	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[expire key %s successed] [expire:%d] [active nums:%d]",
			key, timeout, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Info("[expire key:%s failed] [expire:%d] [error:%s] [active nums:%d]",
		key, timeout, err, r._redis._pool.ActiveCount())
	return 0, errors.New(conf.ERROR_GET_CACHE)
}

func (r *Redis) Expire(key string, timeout int) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.expire(key, timeout)
		if err == nil {
			return v, nil
		}
	}
	r.Warn("[expire key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) del(key string) error {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("DEL", key)
	if err != nil {
		r.Warn("[do del failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return errors.New(conf.ERROR_SET_CACHE)
	}

	_, err = redis.Int(res, err)
	if err == nil {
		r.Debug("[del key %s successed] [active nums:%d]",
			key, r._redis._pool.ActiveCount())
		return nil
	}

	r.Info("[del key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Del(key string) error {
	if r._redis == nil {
		return errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	for i := 0; i < conf.RETRY; i++ {
		err = r.del(key)
		if err == nil {
			return nil
		}
	}
	r.Warn("[del key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return err
}

func (r *Redis) incrBy(key string, value int) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("INCRBY", key, value)
	if err != nil {
		r.Warn("[do incrby failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[incrby key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Info("[incr key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) IncrBy(key string, value int) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.incrBy(key, value)
		if err == nil {
			return v, nil
		}
	}
	r.Warn("[incrby key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) incr(key string) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Critical("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("INCR", key)
	if err != nil {
		r.Warn("[do incr failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[incr key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Info("[incr key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Incr(key string) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.incr(key)
		if err == nil {
			return v, nil
		}
	}
	r.Warn("[incr key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) ttl(key string) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("TTL", key)
	if err != nil {
		r.Warn("[do ttl failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[ttl key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[ttl key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Ttl(key string) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.ttl(key)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[ttl key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) rpush(key string, value ...interface{}) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("rpush", append([]interface{}{key}, value...)...)
	if err != nil {
		r.Warn("[do rpush failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[rpush key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[rpush key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Rpush(key string, value ...interface{}) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.rpush(key, value...)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[rpush key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) lrange(key string, start int, end int) ([]string, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return nil, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("lrange", key, start, end)
	if err != nil {
		r.Warn("[do lrange failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return nil, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Strings(res, err)
	if err == nil {
		r.Debug("[lrange key %s successed] [value:%+v] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[lrange key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return nil, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Lrange(key string, start int, end int) ([]string, error) {
	if r._redis == nil {
		return nil, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v []string
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.lrange(key, start, end)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[lrange key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return nil, err
}

func (r *Redis) script(keyCount int, luaScript string, keysAndArgs ...interface{}) (interface{}, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	s := redis.NewScript(keyCount, luaScript)
	v, err := s.Do(conn, keysAndArgs...)
	if err != nil {
		r.Warn("[do script failed, reconnect] [luaScript:%s] [args:%+v] [error:%s] [active nums:%d]",
			luaScript, keysAndArgs, err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return nil, errors.New(conf.ERROR_SET_CACHE)
	}
	return v, nil
}

func (r *Redis) Script(keyCount int, luaScript string, keysAndArgs ...interface{}) (interface{}, error) {
	if r._redis == nil {
		return nil, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v interface{}
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.script(keyCount, luaScript, keysAndArgs...)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[script:%s failed] [error:%s] [active nums:%d]",
		luaScript, err, r._redis._pool.ActiveCount())
	return nil, err
}

func (r *Redis) hset(key string, field string, value string) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("HSET", key, field, value)
	if err != nil {
		r.Warn("[do hset failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[hset key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[hset key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Hset(key string, field string, value string) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.hset(key, field, value)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[hset key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}

func (r *Redis) hget(key string, field string) (string, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return "", errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("HGET", key, field)
	if err != nil {
		r.Warn("[do hget failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return "", errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.String(res, err)
	if err == nil {
		r.Debug("[hget key %s successed] [value:%s] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[hget key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return "", errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Hget(key string, field string) (string, error) {
	if r._redis == nil {
		return "", errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v string
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.hget(key, field)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[hget key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return "", err
}

func (r *Redis) hgetall(key string) ([]string, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return nil, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("HGETALL", key)
	if err != nil {
		r.Warn("[do hgetall failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return nil, errors.New(conf.ERROR_GET_CACHE)
	}

	v, err := redis.Strings(res, err)
	if err == nil {
		r.Debug("[hgetall key %s successed] [value:%s] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[hgetall key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return nil, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Hgetall(key string) ([]string, error) {
	if r._redis == nil {
		return nil, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v []string
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.hgetall(key)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[hgetall key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return nil, err
}

func (r *Redis) hdel(key string, field string) (int, error) {
	r.StatusStart()
	defer r.StatusEnd()

	conn := r._redis._pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		r.Warn("[get connection failed] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	res, err := conn.Do("HDEL", key, field)
	if err != nil {
		r.Warn("[do hdel failed, reconnect] [error:%s] [active nums:%d]",
			err, r._redis._pool.ActiveCount())
		r.createNamePool(r._name, true)
		return -1, errors.New(conf.ERROR_SET_CACHE)
	}

	v, err := redis.Int(res, err)
	if err == nil {
		r.Debug("[hdel key %s successed] [value:%d] [active nums:%d]",
			key, v, r._redis._pool.ActiveCount())
		return v, nil
	}

	r.Warn("[hdel key %s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, errors.New(conf.ERROR_SET_CACHE)
}

func (r *Redis) Hdel(key string, field string) (int, error) {
	if r._redis == nil {
		return -1, errors.New(conf.ERROR_CONN_CACHE)
	}
	var err error
	var v int
	for i := 0; i < conf.RETRY; i++ {
		v, err = r.hdel(key, field)
		if err == nil {
			return v, nil
		}
	}
	r.Critical("[hdel key:%s failed] [error:%s] [active nums:%d]",
		key, err, r._redis._pool.ActiveCount())
	return -1, err
}
