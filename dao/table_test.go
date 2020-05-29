package dao

import (
	"os"
	"testing"

	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/lib/db"
	"github.com/neil-peng/gomvc/utils"
)

var tC *utils.Context

func TestGetId(t *testing.T) {
}

func TestMain(m *testing.M) {
	utils.SetLogFile(conf.ApiConf.LOG_FILE_NAME)
	utils.SetLogLevel(utils.LEVEL_DEBUG)
	db.Init(&utils.IpServer)
	tC = &utils.Context{
		Logger: utils.NewLogger(),
	}
	os.Exit(m.Run())
}
