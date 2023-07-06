package tools

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

func LoggerInitLevelTag(loggerPath, tag string, configLevel *zapcore.Level) (logger *zap.Logger) {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("init log file ")
		logger, _ = zap.NewDevelopment()
		return logger
	}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "m",
		LevelKey:    "l",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "t",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "f",
		NameKey:      "n",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	//debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	//	return lvl < zapcore.InfoLevel
	//})
	//var configLevel zapcore.Level
	//configLevel, err = zapcore.ParseLevel(logLevel)
	//// 未设置默认warn
	//if err != nil {
	//	configLevel = zapcore.WarnLevel
	//}

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= *configLevel && lvl != zapcore.DebugLevel
	})
	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= *configLevel && lvl == zapcore.DebugLevel
	})

	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	debugWriter := getWriter(loggerPath, "debug_"+tag, tag+"_debug.log")
	infoWriter := getWriter(loggerPath, "info_"+tag, tag+".log")
	//warnWriter := getWriter(loggerPath, "error", "QZ_error.log")

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(debugWriter), debugLevel),
	)

	logger = zap.New(core, zap.AddCaller()) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数
	return logger
}

func getWriter(path, tag, filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		path+"/"+tag+"_%Y%m%d%H.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(fmt.Sprintf("%s/%s", path, filename)),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour),
	)

	if err != nil {
		panic(err)
	}
	return hook
}
