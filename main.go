package main

import (
	"auto_update_ipv6/tools"
	"flag"
	"fmt"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
)

var runOpt = flag.Bool("bg", false, "后台运行")
var configPath = flag.String("c", "./config.yaml", "配置文件")
var stdOutFile = flag.String("stdout", "auto_dns.out", "stdout 重定向")

func main() {
	flag.Parse()

	if *runOpt {
		// 启动一个进程

		// 创建新的进程启动

		//runExec := os.Args[0]

		//fmt.Println("run exec ", runExec)

		//attr := &os.ProcAttr{Env: os.Environ(), Files: []*os.File{nil, nil, nil}}

		err := tools.Background(*stdOutFile, *configPath)

		if err != nil {
			log.Fatalln("启动失败", err)
		}

		// 启动成功
		fmt.Println("启动成功")

		return
	}

	config, err := NewConfig(*configPath)
	//
	if err != nil {
		log.Fatalln("配置文件错误 ", err)
	}

	RunApp(config)

}

func RunApp(config *Config) {
	// 生成客户端

	configLevel, err := zapcore.ParseLevel(config.LoggerLevel)
	// 未设置默认warn
	if err != nil {
		configLevel = zapcore.InfoLevel
	}

	logger := tools.LoggerInitLevelTag(config.LoggerPath, "auto_dns", &configLevel)

	if config.AccessKeyId == "" || config.AccessKeySecret == "" || config.Endpoint == "" {
		logger.Error("没有配置api key 等信息")
		return
	}

	logger.Error("启动参数",
		zap.Any("params", config),
	)

	for true {
		time.Sleep(time.Second * 5)
		runMainApp(logger, config)
		time.Sleep(time.Second * time.Duration(config.RetryTime))
	}

	//if err != nil {
	//	return
	//}

}

func runMainApp(logger *zap.Logger, config *Config) {
	logger.Info("开始任务...")
	defer logger.Info("任务未知情况退出")
	aliClient, errCreateClient := CreateClient(&config.AccessKeyId, &config.AccessKeySecret, config.Endpoint)

	if errCreateClient != nil {
		logger.Error("创建客户端错误", zap.Error(errCreateClient))
		return
	}

	for true {
		// 循环读取配置
		time.Sleep(time.Second * time.Duration(config.UpdateWaitTime))

		// 每分钟探测一次

		updateDnsInfo(logger, config, aliClient)

	}

}

func updateDnsInfo(logger *zap.Logger, config *Config, client *alidns20150109.Client) {
	// 解析本地的ip 信息

	ipv6List := tools.GetTargetIPv6Info(config.InterfaceName)

	if len(ipv6List) == 0 {
		logger.Info("没有读取到ipv6 信息")
		return
	}

	selectTargetIpv6 := tools.SelectIpV6TargetInfo(ipv6List)

	if selectTargetIpv6 == nil {
		logger.Info("没有可用的条件的ipv6")
		return
	}

	if config.LastIpV6Info == selectTargetIpv6.Addr {

		logger.Debug("没有变化不需要更新")
		return
	}

	config.LastIpV6Info = selectTargetIpv6.Addr

	// 更新域名

	updateIpv6ToAlidns(logger, config, client, selectTargetIpv6.Addr)

}

func updateIpv6ToAlidns(logger *zap.Logger, config *Config, client *alidns20150109.Client, addr string) {

	for _, record := range config.Records {

		reqUpdateDomainRecordParams := &alidns20150109.UpdateDomainRecordRequest{}
		reqUpdateDomainRecordParams.SetRecordId(record.RecordId)
		reqUpdateDomainRecordParams.SetRR(record.RR)
		reqUpdateDomainRecordParams.SetValue(addr)
		reqUpdateDomainRecordParams.SetType(record.Type)

		runtimeCtx := &util.RuntimeOptions{}

		tryErr := func() (_err error) {
			defer func() {
				if r := tea.Recover(recover()); r != nil {
					_err = r
				}
			}()

			result, errUpdate := client.UpdateDomainRecordWithOptions(reqUpdateDomainRecordParams, runtimeCtx)

			if errUpdate != nil {
				return errUpdate
			}

			// 解析结果

			logger.Info("请求成功", zap.Any("info", result))

			return nil

		}()

		if tryErr != nil {
			var err = &tea.SDKError{}
			if _t, ok := tryErr.(*tea.SDKError); ok {
				err = _t
			} else {
				err.Message = tea.String(tryErr.Error())
			}
			// 如有需要，请打印 error
			_, _err := util.AssertAsString(err.Message)
			if _err != nil {
				logger.Info("更新错误", zap.Error(_err))
			}
		}
	}
	logger.Info("更新结束")

	return

}

type DomainRecordInfo struct {
	RR       string `json:"RR,omitempty" xml:"RR,omitempty" mapstructure:"RR"`
	RecordId string `json:"RecordId,omitempty" xml:"RecordId,omitempty" mapstructure:"RecordId"`
	Type     string `json:"Type,omitempty" xml:"Type,omitempty" mapstructure:"Type"`
	//Value    *string `json:"Value,omitempty" xml:"Value,omitempty" mapstructure:"Value"`
}

type Config struct {
	LoggerPath      string              `json:"logger_path"    yaml:"logger_path"    mapstructure:"logger_path"`
	LoggerLevel     string              `json:"logger_level"   yaml:"logger_level"   mapstructure:"logger_level"`
	AccessKeyId     string              `json:"accessKeyId,omitempty" xml:"accessKeyId,omitempty" mapstructure:"access_key_id"`
	AccessKeySecret string              `json:"accessKeySecret,omitempty" xml:"accessKeySecret,omitempty" mapstructure:"access_key_secret"`
	Endpoint        string              `json:"endpoint" mapstructure:"endpoint"`
	InterfaceName   string              `json:"net_name" mapstructure:"net_name"`
	LastIpV6Info    string              `json:"-"`
	Records         []*DomainRecordInfo `json:"records" mapstructure:"records"`
	UpdateWaitTime  int                 `json:"update_wait_time" mapstructure:"update_wait_time"`
	RetryTime       int                 `json:"retry_time" mapstructure:"retry_time"`
}

func NewConfig(path string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigFile(path)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := new(Config)

	err = v.UnmarshalKey("ddns", config)

	if err != nil {
		//log.Fatalln("配置文件错误 ", err)
		return nil, err
	}

	// 默认值设置
	if config.LoggerPath == "" {
		config.LoggerPath = "./logs"
	}

	if config.LoggerLevel == "" {
		config.LoggerLevel = "info"
	}

	return config, nil
}

func CreateClient(accessKeyId *string, accessKeySecret *string, endpoint string) (_result *alidns20150109.Client, _err error) {
	//return nil, nil

	config := &openapi.Config{
		// 您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String(endpoint)
	_result = &alidns20150109.Client{}
	_result, _err = alidns20150109.NewClient(config)
	return _result, _err
}
