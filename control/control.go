package control

import (
	"auto_update_ipv6/api"
	"auto_update_ipv6/codes"
	"auto_update_ipv6/tools"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
	"time"
)

func RunControlCmd(unixPath string, reload, store, add, del, show, update, updateLevel bool) {

	esixt, _ := tools.FileExist(codes.AppUnixSockFileName)

	if !esixt {
		// 后台启动 程序

		runErr := tools.Background("mgr.out", "config.yaml")
		if runErr != nil {
			log.Println("启动错误")
			return
		}

	}

	// 连接 rpc

	currentPath, _ := os.Getwd()
	targetUrl := fmt.Sprintf("unix://%s/%s", currentPath, codes.AppUnixSockFileName)

	if unixPath != "" {
		targetUrl = unixPath
	}

	var rpcClient *grpc.ClientConn = nil

	var err error
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		rpcClient, err = grpc.Dial(targetUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		//rpcClient, err = grpc.Dial(targetUrl, grpc.WithInsecure())

		if err != nil {
			log.Println("链接中...")
			continue
		}
		log.Println("链接成功")
		break
	}

	// 进入 主流程

	RunClient(rpcClient, reload, store, add, del, show, update, updateLevel)

}

func RunClient(client *grpc.ClientConn, reload, store, add, del, show, update, updateLevel bool) {

	controlApi := api.NewManagerClient(client)

	defer func() {

		if store {
			StoreConfig(controlApi)
		}
	}()

	if reload {
		ReloadConfig(controlApi)
		return
	}

	if add {
		RunAddRecord(controlApi)
		return
	}

	if show {
		ShowListRecord(controlApi)
		return
	}

	if del {
		DelRecord(controlApi)
		return
	}

	if update {
		fmt.Println("暂不支持")
		return
	}

	if updateLevel {
		updateLoggerLevel(controlApi)
		return
	}
	//if reload {
	//	//
	//}

}

func updateLoggerLevel(controlApi api.ManagerClient) {
	fmt.Println("设置日志等级:")
	fmt.Println("0. debug")
	fmt.Println("1. info")
	fmt.Println("2. warning")
	fmt.Println("3. error")

	inputReader := bufio.NewReader(os.Stdin)

	ctx := context.TODO()

	updateReq := &api.UpdateLoggerLevelReq{}

	for i := 0; i < 3; i++ {
		fmt.Print("输入 ( 0 ~ 3):")

		inputRecordRaw, _ := inputReader.ReadString('\n')
		inputRecord := strings.TrimSpace(inputRecordRaw)
		if inputRecord == "" {
			continue
		}

		switch inputRecord {
		case "0":
			updateReq.Level = api.UpdateLoggerLevelReq_Debug
		case "1":
			updateReq.Level = api.UpdateLoggerLevelReq_Info
		case "2":
			updateReq.Level = api.UpdateLoggerLevelReq_Warning
		case "3":
			updateReq.Level = api.UpdateLoggerLevelReq_Error

		default:
			continue
		}
		break

	}

	_, err := controlApi.UpdateLoggerLevel(ctx, updateReq)

	if err != nil {
		fmt.Println("请求设置日志等级错误", err)
		return
	}

	fmt.Println("设置成功")
	return

}

func ReloadConfig(controlApi api.ManagerClient) {
	ctx := context.TODO()

	writeReq := api.LocalConfigFileReq{}

	_, err := controlApi.LocalConfigFile(ctx, &writeReq)
	if err != nil {
		fmt.Println("读取失败", err)
		return
	}

	fmt.Println("读取成功")
}

func StoreConfig(controlApi api.ManagerClient) {
	ctx := context.TODO()

	writeReq := api.WriteConfigFileReq{}

	_, err := controlApi.WriteConfigFile(ctx, &writeReq)
	if err != nil {
		fmt.Println("保存失败", err)
		return
	}

	fmt.Println("保存成功")

}

func DelRecord(controlApi api.ManagerClient) {
	ctx := context.TODO()

	var delRecord api.DelRecordReq

	inputReader := bufio.NewReader(os.Stdin)

	for i := 0; i < 3; i++ {
		fmt.Print("输入RecordID:")
		inputRecordRaw, _ := inputReader.ReadString('\n')
		inputRecord := strings.TrimSpace(inputRecordRaw)
		if inputRecord == "" {
			continue
		}
		delRecord.RecordId = inputRecord
		break
	}

	if delRecord.RecordId == "" {
		log.Fatal("RecordId 不能为空")
		return
	}

	_, err := controlApi.DelRecord(ctx, &delRecord)

	if err != nil {
		fmt.Println("请求删除错误", err)
		return
	}

	fmt.Println("删除成功")
	return
}

func ShowListRecord(client api.ManagerClient) {
	ctx := context.TODO()

	params := &api.ListRecordReq{
		Limit: 9999,
	}
	recordList, err := client.ListRecord(ctx, params)

	if err != nil {
		log.Println("请求错误", err)
		return
	}

	fmt.Printf("%s:\t%18s\t%6s\t%8s\t%16s\t%s\n", "序号", "RecordID", "RR", "Type", "WatchType", "VMName")
	fmt.Println("-----------------------------------------------------------------------------------------------")
	for i, record := range recordList.Records {
		fmt.Printf("%04d:\t%18s\t%6s\t%8s\t%16s\t%s\n", i, record.RecordId, record.RR, record.Type, record.WatchType.String(), record.VMName)
	}
	fmt.Println("-----------------------------------------------------------------------------------------------")

	fmt.Println("记录总数:", recordList.Count)
}

func RunAddRecord(client api.ManagerClient) {
	// 只支持 ipv6 所以只要选择 类型 和 id 和 RR 即可
	ctx := context.TODO()
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("注意:只支持IPv6 的记录变更")

	var addRecord api.RecordInfo

	addRecord.Type = "AAAA"

	fmt.Print("IP来源ESXI? (Y / N ) [defined N]:")

	inputStringRaw, _ := inputReader.ReadString('\n')
	inputString := strings.TrimSpace(inputStringRaw)
	if inputString == "Y" || inputString == "y" || inputString == "yes" {
		addRecord.WatchType = api.RecordWatchType_WatchTypeEsxi
	}

	switch addRecord.WatchType {
	case api.RecordWatchType_WatchTypeEsxi:
		// 需要设置 主机名
		for i := 0; i < 3; i++ {
			fmt.Print("*输入ESXI实例名(不能为空,VMName):")
			inputVMNameRaw, _ := inputReader.ReadString('\n')
			inputVMName := strings.TrimSpace(inputVMNameRaw)
			if inputVMName == "" {
				continue
			}
			addRecord.VMName = inputVMName
			break
		}
		if addRecord.VMName == "" {
			log.Fatal("错误:实例名为空")
			return
		}
	}

	for i := 0; i < 3; i++ {
		fmt.Print("输入RecordID:")
		inputRecordRaw, _ := inputReader.ReadString('\n')
		inputRecord := strings.TrimSpace(inputRecordRaw)
		if inputRecord == "" {
			continue
		}
		addRecord.RecordId = inputRecord
		break
	}

	if addRecord.RecordId == "" {
		log.Fatal("RecordId 不能为空")
		return
	}

	for i := 0; i < 3; i++ {
		fmt.Print("输入RecordID的RR值:")
		inputRRRaw, _ := inputReader.ReadString('\n')
		inputRR := strings.TrimSpace(inputRRRaw)
		if inputRR == "" {
			continue
		}
		addRecord.RR = inputRR
		break
	}

	if addRecord.RR == "" {
		log.Fatal("RR 记录不能为空")
		return
	}

	fmt.Println("是否提交一下记录:")

	fmt.Println("RR:", addRecord.RR)
	fmt.Println("RecordId:", addRecord.RecordId)
	fmt.Println("Type:", addRecord.Type)
	fmt.Println("WatchType:", addRecord.WatchType)
	fmt.Println("VMName:", addRecord.VMName)

	inputYesRaw, _ := inputReader.ReadString('\n')

	inputYes := strings.TrimSpace(inputYesRaw)

	if inputYes == "Y" || inputYes == "y" || inputYes == "yes" {

		_, err := client.AddRecord(ctx, &addRecord)

		if err != nil {
			fmt.Println("添加请求错误", err)
			return
		}

		fmt.Println("添加成功")
		return
	} else {
		fmt.Println("放弃了请求")
		return
	}

}
