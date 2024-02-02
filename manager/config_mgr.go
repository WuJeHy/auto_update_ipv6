package manager

import (
	"context"
	"github.com/spf13/viper"
	"log"
	"time"
)

// 本地的配置文件的映射
var localConfig = new(ConfigFile)

// 配置管理器
var viperObject = viper.New()

func InitConfig(path string) error {
	viperObject.AddConfigPath(".")
	viperObject.SetConfigFile(path)
	return LoadConfigFile()
}

func LoadConfigFile() (err error) {
	err = viperObject.ReadInConfig()

	if err != nil {
		return
	}

	// 读取成功解析到 localConfig

	err = viperObject.Unmarshal(localConfig)
	if err != nil {
		return
	}

	return err
}

// 将配置写入到文件
//func WriteConfig() error {
//
//	viperObject.WriteConfig()
//
//}

func RunTestConfig(path string) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigType("yaml")
	//v.SetConfigFile(path)
	//v.SetConfigFile("test2.yaml")

	err := v.ReadInConfig()
	if err != nil {
		log.Println("fail exit ", err)
		return
	}
	//loadConfig := &ConfigFile{}
	fileConfig := &ConfigFile{}

	errRead := v.Unmarshal(fileConfig)
	if errRead != nil {
		log.Println("err read ", errRead)
	}

	//v.WatchConfig()

	//fileConfig := &ConfigFile{}

	fileConfig.Logger.LoggerLevel = "debug"
	fileConfig.Logger.LoggerPath = "logs"

	v.Set("logger", &fileConfig.Logger)
	if err != nil {
		log.Println("test add ", err)
	}
	fileConfig.Logger.LoggerPath = "logsa"
	// 配置更新的事件
	//v.OnConfigChange(func(in fsnotify.Event) {
	//	log.Println("show event ", in.Name, in.Op.String())
	//	// 这里 查看到
	//})

	v.MergeInConfig()
	//err = v.WriteConfig()
	//if err != nil {
	//	log.Println("write ", err)
	//}

	ctx := context.Background()

	ctxApp, ctxFunc := context.WithTimeout(ctx, time.Minute)

	defer ctxFunc()

	<-ctxApp.Done()

	log.Println("exit")

}

// 处理管理器
