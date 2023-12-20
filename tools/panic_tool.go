package tools

import "go.uber.org/zap"

func HandlePanic(logger *zap.Logger, tag string) {
	// 未知原因的错误处理
	// 捕捉错误
	errRecover := recover()
	if errRecover == nil {
		// 没有错误
		return
	}

	//err = errRecover
	// 有错误 直接打印错误
	logger.Error(tag, zap.Any("err", errRecover))
}
