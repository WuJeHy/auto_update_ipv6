package manager

import (
	"auto_update_ipv6/tools"
	"time"
)

func (m *Manager) RunUpdateServices() {
	defer tools.HandlePanic(m.logger, "RunUpdateServices")
	// 执行 更新服务
	logger := m.logger.Named("RunUpdateServices")

	// 自动更新延时

	UpdateWaitTime := func() int {
		defer tools.HandlePanic(logger, "GetWaitTimePanic")
		return m.fileConfig.App.UpdateWaitTime
	}()

	// 默认 120s 触发一次
	if UpdateWaitTime == 0 {
		UpdateWaitTime = 120
	}

	ticker := time.NewTicker(time.Second * time.Duration(UpdateWaitTime))

	// 定时任务
	for true {
		select {
		case <-m.ctx.Done():
			logger.Info("管理器结束任务")
			return
		case timeNow := <-ticker.C:
			// ip 信息
			m.doGetIpAddrUpdate(timeNow)

		}
	}

}
