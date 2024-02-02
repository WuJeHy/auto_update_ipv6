package main

import (
	"auto_update_ipv6/control"
	"auto_update_ipv6/manager"
	"auto_update_ipv6/tools"
	//"auto_update_ipv6/ui"
	"flag"
	"fmt"
	"log"
)

var startUI = flag.Bool("ui", false, "启动ui")
var runOpt = flag.Bool("bg", false, "后台运行")
var configPath = flag.String("c", "./config.yaml", "配置文件")
var stdOutFile = flag.String("stdout", "mgr.out", "stdout 重定向")
var unixSockPath = flag.String("unix", "", "unix sock 文件 路径")

var reloadFlag = flag.Bool("reload", false, "重载记录配置")
var storeFlag = flag.Bool("store", false, "保存记录配置")
var addRecord = flag.Bool("add", false, "添加一个记录")
var delRecord = flag.Bool("del", false, "删除一个记录")
var showRecord = flag.Bool("show", false, "查看记录")
var updateRecord = flag.Bool("update", false, "更新记录")
var updateLoggerLevel = flag.Bool("logger_level", false, "更新日志等级")

func main() {

	flag.Parse()
	if *runOpt {
		// 启动一个进程

		// 创建新的进程启动

		err := tools.Background(*stdOutFile, *configPath)

		if err != nil {
			log.Fatalln("启动失败", err)
		}

		// 启动成功
		fmt.Println("启动成功")

		return
	}
	// todo 交互 ui
	//if *startUI {
	//	ui.Run(*unixSockPath)
	//	return
	//}

	if *reloadFlag ||
		*storeFlag ||
		*addRecord ||
		*delRecord ||
		*showRecord ||
		*updateRecord ||
		*updateLoggerLevel {
		reload := *reloadFlag
		store := *storeFlag
		add := *addRecord
		del := *delRecord
		show := *showRecord
		update := *updateRecord
		updateLevel := *updateLoggerLevel
		control.RunControlCmd(*unixSockPath, reload, store, add, del, show, update, updateLevel)
		return
	}

	err := manager.RunManagerApp(*configPath)
	if err != nil {
		log.Println("err ", err)
	}
}
