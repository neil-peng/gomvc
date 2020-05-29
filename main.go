package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/neil-peng/gomvc/action"
	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/lib/db"
	"github.com/neil-peng/gomvc/utils"
)

func setupSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGHUP)
	signal.Notify(c, syscall.SIGUSR1)
	signal.Notify(c, syscall.SIGUSR2)
	signal.Notify(c, syscall.SIGTTIN)
	signal.Notify(c, syscall.SIGTTOU)
	signal.Notify(c, syscall.SIGPIPE)
	go func() {
		for sig := range c {
			utils.Warn("got sig:%v", sig)
			switch sig {
			case syscall.SIGUSR1:
				f, err := os.Create(conf.ApiConf.LOG_FILE_DIR + " gomvc.prof")
				if err != nil {
					utils.Warn("create gomvc.prof erorr")
					break
				}
				utils.Notice("start cpu pprof")
				pprof.StartCPUProfile(f)
			case syscall.SIGUSR2:
				utils.Notice("stop cpu pprof")
				pprof.StopCPUProfile()
			case syscall.SIGTTIN:
				utils.SetLogLevel(utils.GetLogLevel() + 1)
			case syscall.SIGTTOU:
				utils.SetLogLevel(utils.GetLogLevel() - 1)
			case syscall.SIGHUP:
				utils.Info("SIGHUP")
				utils.ReOpen("")
			case syscall.SIGPIPE:
				utils.Warn("ignore sig:%v", sig)
			case syscall.SIGINT:
				utils.Warn("SIGINT")
				os.Exit(0)
			}
		}
	}()
}

func Init() {
	setupSignal()
	utils.SetLogFile(conf.ApiConf.LOG_FILE_NAME)
	utils.SetLogLevel(conf.ApiConf.LOG_LEVEL)
	utils.SetLogbackupCount(48) //live: 2 days
	utils.SetLogRotate(time.Hour)
	db.Init(&utils.IpServer)
}

func main() {
	Init()
	utils.AddRoute("GET", "/rest/example/add", &action.Api{}, action.AddExample)
	utils.AddRoute("GET", "/rest/example/get", &action.Api{}, action.GetExample)
	utils.RunServer(":8080")
}
