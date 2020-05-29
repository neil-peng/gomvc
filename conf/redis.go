package conf

const (
	REDIS_CONNECT_TIMEOUTMS        = 100
	REDIS_READ_TIMEOUTMS           = 100
	REDIS_WRITE_TIMEOUTMS          = 100
	REDIS_POOL_INT_MAX_IDLE_NUMS   = 100
	REDIS_POOL_INT_MAX_ACTIVE_NUMS = 100
	REDIS_POOL_INT_IDLE_TIMEOUT    = 100
	REDIS_POOL_CONN_KEEPALIVE_NUMS = 500
	REDIS_POOL_BOOL_WAIT           = false
)

var RedisInfo = map[string]interface{}{
	"key": map[string]interface{}{
		"prefix":       "gomvc-pre-",
		"expire":       60,
		"defaultvalue": "yes",
		"addr":         "",
	},
}
