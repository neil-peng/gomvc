package conf

import (
	"strconv"
)

const (
	NO_ERROR = "0"
	//db error
	ERROR_DB_QUERY_ERROR      = "10001"
	ERROR_DB_QUERY_DUPLICATE  = "10000"
	ERROR_DB_CONNECT_ERROR    = "10002"
	ERROR_DB_RESULT_SET_EMPTY = "10003"
	//system error
	ERROR_NETWORK_ERROR        = "10004"
	ERROR_SERVER_NOT_ACCESS    = "10005"
	ERROR_PARAM_ERROR          = "10006"
	ERROR_INNER_ERROR          = "10007"
	ERROR_NAMESERVICE_ERROR    = "10008"
	ERROR_CONF_ERROR           = "10009"
	ERROR_CONN_CACHE           = "10010"
	ERROR_GET_CACHE            = "10011"
	ERROR_SET_CACHE            = "10012"
	ERROR_FIELD_SCHEME_INVALID = "10013"
)

var ArrErrorMessage = map[string]string{
	NO_ERROR:                   "success",
	ERROR_DB_QUERY_ERROR:       "db query error",
	ERROR_DB_QUERY_DUPLICATE:   "db duplicate entry error",
	ERROR_DB_CONNECT_ERROR:     "db connect error",
	ERROR_DB_RESULT_SET_EMPTY:  "db result set is empty",
	ERROR_NETWORK_ERROR:        "network error",
	ERROR_SERVER_NOT_ACCESS:    "can not access server",
	ERROR_PARAM_ERROR:          "param error",
	ERROR_INNER_ERROR:          "innner error",
	ERROR_NAMESERVICE_ERROR:    "get nameservice error",
	ERROR_CONF_ERROR:           "parse conf error",
	ERROR_CONN_CACHE:           "connect cache error",
	ERROR_GET_CACHE:            "get cache error",
	ERROR_SET_CACHE:            "set cache error",
	ERROR_FIELD_SCHEME_INVALID: "db scheme error",
}

var ArrHttpCode = map[string]int{
	NO_ERROR:                   200,
	ERROR_DB_QUERY_ERROR:       503,
	ERROR_DB_QUERY_DUPLICATE:   503,
	ERROR_DB_CONNECT_ERROR:     503,
	ERROR_DB_RESULT_SET_EMPTY:  503,
	ERROR_NETWORK_ERROR:        503,
	ERROR_SERVER_NOT_ACCESS:    503,
	ERROR_PARAM_ERROR:          400,
	ERROR_INNER_ERROR:          503,
	ERROR_NAMESERVICE_ERROR:    503,
	ERROR_CONF_ERROR:           503,
	ERROR_CONN_CACHE:           503,
	ERROR_GET_CACHE:            503,
	ERROR_SET_CACHE:            503,
	ERROR_FIELD_SCHEME_INVALID: 503,
}

func GetHttpCode(errno string) int {
	if _, ok := ArrHttpCode[errno]; ok {
		return ArrHttpCode[errno]
	}
	errorInt, _ := strconv.Atoi(errno)
	if errorInt < 1000 {
		return 401
	}
	return 200
}

func GetHttpMsg(errno string) string {
	if _, ok := ArrErrorMessage[errno]; ok {
		return ArrErrorMessage[errno]
	}
	return "unknow errno:" + errno
}
