package conf

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type Conf_Api struct {
	PORT          int
	LOG_LEVEL     int32
	LOG_FILE_NAME string
	LOG_FILE_DIR  string
}

const RETRY = 3

var ApiConf Conf_Api
var Db Conf_Db

//配置初始化
func init() {
	appPath := os.Getenv("APP_PATH")
	appPath = "./"
	if _, err := toml.DecodeFile(appPath+"/conf/gomvc.toml", &ApiConf); err != nil {
		panic(err)
	}

	if _, err := toml.DecodeFile(appPath+"/conf/db.toml", &Db); err != nil {
		panic(err)
	}

	tableViewArray := Db.Table_view
	tableDbTagMap = make(map[string]string)
	for _, tableView := range tableViewArray {
		tableDbTagMap[tableView.Table_name] = tableView.Db_cluster_tag
	}

	ApiConf.LOG_FILE_DIR = appPath + "/log/"
	ApiConf.LOG_FILE_NAME = ApiConf.LOG_FILE_DIR + ApiConf.LOG_FILE_NAME
	return
}

func V(item map[string]interface{}, keys ...string) (value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panic(errors.New(ERROR_CONF_ERROR))
		}
	}()
	if len(keys) == 0 {
		return item
	}
	rootKey := keys[0]
	if rootIns, ok := (item[rootKey]).(map[string]interface{}); ok {
		if len(keys) > 1 {
			otherKeys := keys[1:]
			return V(rootIns, otherKeys...)
		}
	}
	value = item[rootKey]
	return
}

func Bool(item interface{}, keys ...string) bool {
	if itemMI, ok := item.(map[string]interface{}); ok {
		r := V(itemMI, keys...)
		if rB, ok := r.(bool); ok {
			return rB
		} else if rM, ok := r.(map[string]bool); ok {
			if len(keys) > 0 {
				return rM[keys[len(keys)-1]]
			}
		}
		return false
	} else if itemB, ok := item.(bool); ok {
		return itemB
	}
	return false
}

func Int(item interface{}, keys ...string) int {
	if itemMI, ok := item.(map[string]interface{}); ok {
		r := V(itemMI, keys...)
		if rI, ok := r.(int); ok {
			return rI
		} else if rM, ok := r.(map[string]int); ok {
			if len(keys) > 0 {
				//fixme: no exist retrun 0, not -1
				return rM[keys[len(keys)-1]]
			}
		}
		return -1
	} else if itemInt, ok := item.(int); ok {
		return itemInt
	}
	return -1
}

func String(item interface{}, keys ...string) string {
	if itemMI, ok := item.(map[string]interface{}); ok {
		r := V(itemMI, keys...)
		if rS, ok := r.(string); ok {
			return rS
		} else if rM, ok := r.(map[string]string); ok {
			if len(keys) > 0 {
				return rM[keys[len(keys)-1]]
			}
		}
		return ""
	} else if itemB, ok := item.(string); ok {
		return itemB
	}
	return ""
}
