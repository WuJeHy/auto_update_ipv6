package control

import (
	"auto_update_ipv6/api"
	"auto_update_ipv6/tools"
	"context"
	"errors"
	"go.uber.org/zap"
)

func (m *Manager) LocalConfigFile(ctx context.Context, req *api.LocalConfigFileReq) (resp *api.LocalConfigFileResp, err error) {
	// 从配置文件读取
	defer tools.HandlePanic(m.logger, "LocalConfigFile")

	// 读取配置文件的路径 路径在参数中

	logger := m.logger.Named("LocalConfigFile")
	// 读取配置文件
	errRead := m.viper.ReadInConfig()

	if errRead != nil {
		logger.Error("Reader Fail ", zap.Error(errRead))
		err = errors.New("读取配置文件错误")
		return
	}

	unmarshalErr := m.viper.UnmarshalKey("records", &m.recordsConfig.Records)
	if unmarshalErr != nil {
		err = errors.New("解析错误")
		return
	}

	resp = new(api.LocalConfigFileResp)

	return
}

func (m *Manager) WriteConfigFile(ctx context.Context, req *api.WriteConfigFileReq) (resp *api.WriteConfigFileResp, err error) {
	// 将 配置文件 写到文件
	defer tools.HandlePanic(m.logger, "WriteConfigFile")

	logger := m.logger.Named("WriteConfigFile")

	if len(m.recordsConfig.Records) == 0 {
		err = errors.New("没有更新的记录")
		return
	}

	m.viper.Set("records", m.recordsConfig.Records)
	var writeErr error
	if req.GetRecordFile() != "" {
		writeErr = m.viper.WriteConfigAs(req.GetRecordFile())
	} else {
		writeErr = m.viper.WriteConfig()
	}

	if writeErr != nil {
		logger.Warn("更新记录文件错误", zap.Error(writeErr))
		err = errors.New("更新记录文件失败")
		return
	}

	resp = new(api.WriteConfigFileResp)

	return
}

func (m *Manager) AddRecord(ctx context.Context, req *api.RecordInfo) (resp *api.AddRecordResp, err error) {
	defer tools.HandlePanic(m.logger, "AddRecord")

	// 添加一条记录

	exist := m.CheckRecordExistCallback(req.RecordId)

	if exist {
		err = errors.New("记录已存在")
		return
	}

	// 添加

	newRecord := &DomainRecordInfo{
		RecordId: req.RecordId,
		Type:     req.Type,
		RR:       req.RR,
	}

	switch req.GetWatchType() {
	case api.RecordWatchType_WatchTypeEsxi:
		newRecord.WatchType = tools.DomainRecordInfoWatchTypeEsxi

		if req.VMName == "" {
			err = errors.New("esxi 类型下没有配置 VMNAME")
			return
		}
		newRecord.VMName = req.VMName
	default:
		newRecord.WatchType = tools.DomainRecordInfoWatchTypeLocal

	}

	// 添加记录

	m.recordsConfig.Records = append(m.recordsConfig.Records, newRecord)

	resp = new(api.AddRecordResp)

	return
}

func (m *Manager) DelRecord(ctx context.Context, req *api.DelRecordReq) (resp *api.DelRecordResp, err error) {
	defer tools.HandlePanic(m.logger, "DelRecord")

	m.DeleteRecordByRecordID(req.GetRecordId())

	resp = new(api.DelRecordResp)
	return
}

func (m *Manager) EditRecord(ctx context.Context, req *api.EditRecordReq) (resp *api.EditRecordResp, err error) {

	defer tools.HandlePanic(m.logger, "EditRecord")

	exist := m.CheckRecordExistCallback(req.RecordId, func(record *DomainRecordInfo) {
		if req.VMName != nil {
			record.VMName = req.GetVMName()
		}

		if req.WatchType != nil {
			switch req.GetWatchType() {
			case api.RecordWatchType_WatchTypeEsxi:
				record.WatchType = tools.DomainRecordInfoWatchTypeEsxi
			case api.RecordWatchType_WatchTypeLocal:
				record.WatchType = tools.DomainRecordInfoWatchTypeLocal
			}
		}

		if req.RR != nil {
			record.RR = req.GetRR()
		}

		if req.Type != nil {
			record.Type = req.GetType()
		}
	})

	if !exist {
		err = errors.New("记录不存在")
		return
	}

	resp = new(api.EditRecordResp)
	return
}

func (m *Manager) ListRecord(ctx context.Context, req *api.ListRecordReq) (resp *api.ListRecordResp, err error) {

	resp = new(api.ListRecordResp)

	domainRecordInfoTpApiRecordInfo := func(recordInfo *DomainRecordInfo) *api.RecordInfo {
		tmp := &api.RecordInfo{
			RR:       recordInfo.RR,
			RecordId: recordInfo.RecordId,
			Type:     recordInfo.Type,
			VMName:   recordInfo.VMName,
		}

		switch recordInfo.Type {
		case tools.DomainRecordInfoWatchTypeEsxi:
			tmp.WatchType = api.RecordWatchType_WatchTypeEsxi
		default:
			tmp.WatchType = api.RecordWatchType_WatchTypeLocal
		}

		return tmp
	}

	doGetSliceList := func() {

		maxLen := len(m.recordsConfig.Records)

		resp.Count = int64(maxLen)
		if resp.Count < req.Offset {
			// 跳过的过多 直接空了
			return
		}

		endSliceIndex := req.Offset + req.Limit

		var findSlice []*DomainRecordInfo

		if resp.Count < endSliceIndex {
			// 中途就没有了 所以只能切到 offset -> end

			findSlice = m.recordsConfig.Records[req.Offset:]

		} else {
			// 足够长
			findSlice = m.recordsConfig.Records[req.Offset:endSliceIndex]
		}

		// 转化

		for _, recordInfo := range findSlice {
			//tmpAddRespList := recordInfo
			tmp := domainRecordInfoTpApiRecordInfo(recordInfo)
			resp.Records = append(resp.Records, tmp)
		}

	}

	doGetWatchTypeList := func(watchType string) {
		// 遍历全部的

		var count int64 = 0
		var found int64 = 0

		for _, recordInfo := range m.recordsConfig.Records {
			if recordInfo.WatchType == watchType {
				count++

				if found < req.Limit {
					found++
					tmp := domainRecordInfoTpApiRecordInfo(recordInfo)
					resp.Records = append(resp.Records, tmp)
				}
			}
		}

		resp.Count = count

	}

	if req.WatchType == nil {
		// 直接切片最快
		doGetSliceList()
		return
	} else {
		// 一般记录不会太多 使用 for 循环筛选
		switch req.GetWatchType() {
		case api.RecordWatchType_WatchTypeEsxi:
			doGetWatchTypeList(tools.DomainRecordInfoWatchTypeEsxi)
			return
		case api.RecordWatchType_WatchTypeLocal:
			doGetWatchTypeList(tools.DomainRecordInfoWatchTypeLocal)
		default:
			doGetSliceList()
			return
		}
	}

	return
}
