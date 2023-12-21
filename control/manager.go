package control

import (
	"auto_update_ipv6/api"
	"auto_update_ipv6/tools"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 主控制器
//1. 核心更新服务
//a. 附带 sock 接口
//b. 附带 服务上下文
//c. 可以通过接口软重启
//d. 可以配置 yaml
//e. 日志服务

// Manager 主管理服务
type Manager struct {
	configPath string

	// 需要缓存一个 配置文件
	fileConfig *ConfigFile
	// 需要一个 配置管理器
	viper *viper.Viper
	// 需要一个 日志系统
	logger *zap.Logger
	// 需要保证程序 的上下文
	ctx           context.Context
	cancelCtxFunc func()

	rw            *sync.RWMutex
	recordsConfig *RecordsConfig
	//	rpc
	grpcd *grpc.Server
}

func (m *Manager) Ctx() context.Context {
	return m.ctx
}

func (m *Manager) Close() error {

	if m.cancelCtxFunc != nil {
		m.cancelCtxFunc()
	}

	if m.grpcd != nil {
		m.grpcd.GracefulStop()
	}

	return nil
}

//func NewManager(logger *zap.Logger, config *ConfigFile) *Manager {
//
//	ctxTodo := context.TODO()
//
//	return NewManagerWithContext(ctxTodo, logger, config)
//}

func NewManagerWithContext(ctx context.Context, logger *zap.Logger, config *ConfigFile) *Manager {

	ctxApp, ctxFunc := context.WithCancel(ctx)

	v := viper.New()

	recordsConfig := new(RecordsConfig)

	mgr := &Manager{
		ctx:           ctxApp,
		cancelCtxFunc: ctxFunc,
		logger:        logger,
		viper:         v,
		fileConfig:    config,
		recordsConfig: recordsConfig,
		rw:            new(sync.RWMutex),
	}
	if config.RecordsFile == "" {
		mgr.configPath = "records.yaml"
	} else {
		mgr.configPath = config.RecordsFile
	}
	v.SetConfigFile(mgr.configPath)

	return mgr
}

// RunManager 使用配置文件启动
func RunManager(path string) error {
	// 启动运行

	// 解析 配置文件

	// 生成一个初试的配置文件

	// 如果日志文件配置为空 则 进入默认配置

	_, config, err := newConfig(path)

	if err != nil {
		return err
	}
	//初始化 日志系统

	configLevel, err := zapcore.ParseLevel(config.Logger.LoggerLevel)
	// 未设置默认warn
	if err != nil {
		configLevel = zapcore.InfoLevel
	}

	logger := tools.LoggerInitLevelTag(config.Logger.LoggerPath, "ddns", &configLevel)

	if logger == nil {
		return errors.New("日志系统初始化失败")
	}

	ctxBg := context.Background()

	mgr := NewManagerWithContext(ctxBg, logger, config)

	if mgr == nil {
		return errors.New("实例化 管理器失败")
	}

	// 启动服务

	go WatchQuitSignal(logger, mgr)

	mgr.InitGrpc()

	errExit := mgr.Run()

	if errExit != nil {
		return errExit
	}

	return nil
}

func WatchQuitSignal(logger *zap.Logger, mgr interface {
	Close() error
	Ctx() context.Context
}) {

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case <-mgr.Ctx().Done():
	}
	closeErr := mgr.Close()

	logger.Error("Shutdown Server ...", zap.Error(closeErr))

}

func (m *Manager) Run() error {
	// 运行

	// 启动 grpc
	// 使用 unix sock 方式启动 不使用 端口方式

	logger := m.logger
	lis, err := net.Listen("unix", AppUnixSockFileName)

	//lis, err := net.Listen("tcp", ":18411")

	if err != nil {
		return err
	}

	defer lis.Close()

	// 启动 grpc
	// 因为只有一个服务 所以直接以 grpc 的生命周期为主进程进行阻塞
	err = m.grpcd.Serve(lis)
	if err != nil {
		fmt.Printf("开启服务失败: %s", err)
	}

	logger.Info("grpc 运行结束，主进程退出")

	logger.Info("等待其他任务结束")

	waitCloseSleep(logger, 3)

	return nil
}

func waitCloseSleep(logger *zap.Logger, i int) {

	for i > 0 {
		time.Sleep(time.Second)
		logger.Info("关闭倒计时", zap.Int("t", i))
		i--
	}
}

func (m *Manager) InitGrpc() {
	// 初始化配置grpc

	grpcServer := grpc.NewServer()

	m.grpcd = grpcServer

	api.RegisterManagerServer(grpcServer, m)

}

func (m *Manager) CheckRecordExistCallback(recordId string, cbs ...func(record *DomainRecordInfo)) (exist bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	for _, record := range m.recordsConfig.Records {
		if record.RecordId == recordId {
			// 找到
			exist = true
			if len(cbs) != 0 {
				m.rw.RUnlock()
				for _, cb := range cbs {
					cb(record)
				}
				m.rw.RLock()
			}
			break

		}
	}

	return
}

func (m *Manager) DeleteRecordByRecordID(recordId string) {
	m.rw.Lock()
	defer m.rw.Unlock()

	for i, record := range m.recordsConfig.Records {
		if record.RecordId == recordId {
			m.recordsConfig.Records = append(m.recordsConfig.Records[:i], m.recordsConfig.Records[i+1:]...)
			return
		}
	}

}
func newConfig(path string) (v *viper.Viper, config *ConfigFile, err error) {

	v = viper.New()

	v.SetConfigFile(path)

	config = new(ConfigFile)

	err = v.ReadInConfig()
	if err != nil {
		// 文件不存在
		//log.Println("读取配置文件不存在,准备初始化一个空的配置")

		return
	}

	err = v.Unmarshal(config)

	// 默认值设置
	if config.Logger.LoggerPath == "" {
		config.Logger.LoggerPath = "./logs"
	}

	if config.Logger.LoggerLevel == "" {
		config.Logger.LoggerLevel = "info"
	}

	return
}
