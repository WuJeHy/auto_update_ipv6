package manager

import (
	"auto_update_ipv6/tools"
	"errors"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"
	"time"
)

func getCurrentAvailableIpv6(logger *zap.Logger, interfaceName string) string {
	ipv6List := tools.GetTargetIPv6Info(interfaceName)

	if len(ipv6List) == 0 {
		logger.Debug("没有读取到ipv6 信息")
	}

	selectTargetIpv6 := tools.SelectIpV6TargetInfo(ipv6List)

	if selectTargetIpv6 == nil {
		logger.Debug("没有可用的条件的ipv6")
		return ""
	}

	return selectTargetIpv6.Addr
}

func (m *Manager) doGetIpAddrUpdate(timeNow time.Time) {
	// 读取本地 ipv6 信息
	logger := m.logger
	defer tools.HandlePanic(logger, "doGetIpAddrUpdate")

	interfaceName := m.fileConfig.App.InterfaceName
	currentIPv6AddrStr := getCurrentAvailableIpv6(logger, interfaceName)

	// 读取 esxi 信息
	esxiConfig := m.fileConfig.Esxi
	getEsxiIpv6, errGetAllEsxiAddrs := tools.GetAllEsxiAddrs(logger, esxiConfig.Url, esxiConfig.Username, esxiConfig.Password, esxiConfig.Insecure)
	if errGetAllEsxiAddrs != nil {
		logger.Debug("读取esxi 失败", zap.Error(errGetAllEsxiAddrs))
	}

	// update

	aliConfig := &m.fileConfig.Aliyun

	if aliConfig.AccessKeyId == "" || aliConfig.AccessKeySecret == "" {
		logger.Warn("AccessKeyId 或 AccessKeySecret 没有设置")
		return
	}

	aliClient, errCreateAliClient := CreateClient(&aliConfig.AccessKeyId, &aliConfig.AccessKeySecret, aliConfig.Endpoint)

	if errCreateAliClient != nil {
		logger.Debug("创建AliDNSClient fail")
		return
	}

	updateAllIPV6ToAliyunDDns(m, aliClient, currentIPv6AddrStr, getEsxiIpv6)

}

func updateAllIPV6ToAliyunDDns(m *Manager, aliClient *alidns20150109.Client, currentIPv6AddrStr string, esxiIpv6Map map[string]string) {
	// 匹配到需要更新的数据进行更新

	logger := m.logger.Named("UpdateDNSConfig")
	defer tools.HandlePanic(logger, "updateAllIPV6ToAliyunDDns")
	// 读取 AliClient
	updateRecord := func(record *DomainRecordInfo) {
		// 捕捉异常
		// 单次错误不至于导致整体错误
		defer tools.HandlePanic(logger, "updateRecord")
		reqUpdateDomainRecordParams := &alidns20150109.UpdateDomainRecordRequest{}
		reqUpdateDomainRecordParams.SetRecordId(record.RecordId)
		reqUpdateDomainRecordParams.SetRR(record.RR)

		switch record.WatchType {
		case tools.DomainRecordInfoWatchTypeEsxi:
			// 使用 来自 esxi 的记录
			//
			// 匹配记录对应的 ip

			// note 如果 esxi 离线的时候 esxiMap 则空 会导致程序崩溃
			if esxiIpv6Map == nil {
				return
			}

			findAddr, isok := esxiIpv6Map[record.VMName]
			if !isok || findAddr == "" {
				logger.Debug("没有读取到对应实例的ip")
				return
			}

			// 读取到了 对比记录值

			if record.LastVMAddr == findAddr {
				logger.Debug("ip 没有变化", zap.String("vm name", record.VMName), zap.String("last", record.LastVMAddr), zap.String("current", findAddr))
				return
			}

			// 更新最新值
			record.LastVMAddr = findAddr
			reqUpdateDomainRecordParams.SetValue(findAddr)

		default:
			// 默认的是本机的

			if record.LastVMAddr == currentIPv6AddrStr || currentIPv6AddrStr == "" {
				// 不处理
				logger.Debug("本地ip没有变化不需要更新")
				return
			}

			record.LastVMAddr = currentIPv6AddrStr

			reqUpdateDomainRecordParams.SetValue(currentIPv6AddrStr)
		}

		reqUpdateDomainRecordParams.SetType(record.Type)

		runtimeCtx := &util.RuntimeOptions{}

		tryErr := func() (_err error) {

			defer func() {

				if r := tea.Recover(recover()); r != nil {
					_err = r
					logger.Error("unknown errr", zap.Error(_err))
				}
			}()

			logger.Debug("UpdateDomainRecordWithOptions start ", zap.Any("reqUpdateDomainRecordParams", reqUpdateDomainRecordParams))
			if aliClient == nil {
				_err = errors.New("client not init ")
				return
			}

			result, errUpdate := aliClient.UpdateDomainRecordWithOptions(reqUpdateDomainRecordParams, runtimeCtx)
			logger.Debug("UpdateDomainRecordWithOptions end", zap.Any("result", result), zap.Error(errUpdate))

			if errUpdate != nil {
				return errUpdate
			}
			// 解析结果

			logger.Info("请求成功", zap.Any("info", result))

			return nil

		}()

		if tryErr != nil {
			logger.Debug("Error ", zap.Error(tryErr))
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

	// 遍历 处理

	for index, record := range m.recordsConfig.Records {
		logger.Debug("开始处理", zap.Int("index", index), zap.String("ID", record.RecordId))
		updateRecord(record)
		//logger.Debug("处理结束", zap.Int("index", index), zap.String("ID", record.RecordId))
	}

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
