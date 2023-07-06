package tools

import (
	"os/exec"
	"strconv"
	"strings"
)

type Ipv6Info struct {
	Addr      string
	Mark      string
	Scope     []string
	ValidTime int
	Preferred int
}

func GetTargetIPv6Info(name string) []*Ipv6Info {
	//	ip addr show dev wlp5s0  scope global up
	outputCmd := exec.Command("ip", "addr", "show", "dev", name, "scope", "global", "up")

	resultBytes, err := outputCmd.Output()
	if err != nil {
		return nil
	}

	dataString := string(resultBytes)

	findStart := strings.Index(dataString, "inet")

	if findStart < 0 {
		return nil
	}

	canProcStr := dataString[findStart:]
	stripList := strings.Split(canProcStr, "\n")

	dataLen := len(stripList)

	var result []*Ipv6Info
	for i := 0; i < dataLen-1; i = i + 2 {

		ipInfo := strings.TrimSpace(stripList[i])
		extInfo := strings.TrimSpace(stripList[i+1])

		findIpv6Index := strings.Index(ipInfo, "inet6")

		if findIpv6Index < 0 {
			continue
		}

		findScopeIndex := strings.Index(ipInfo, "scope")

		if findScopeIndex < 0 {
			continue
		}

		findMarkIndex := strings.Index(ipInfo, "/")

		if findMarkIndex < 0 {
			continue
		}

		findValidLftIndex := strings.Index(extInfo, "valid_lft")
		findPreferredLftIndex := strings.Index(extInfo, "preferred_lft")

		if findValidLftIndex < 0 {
			continue
		}
		if findPreferredLftIndex < 0 {
			continue
		}

		tmp := &Ipv6Info{}
		ipaddrStr := ipInfo[findIpv6Index+6 : findMarkIndex]
		ipMark := ipInfo[findMarkIndex+1 : findScopeIndex-1]

		tmp.Addr = ipaddrStr
		tmp.Mark = ipMark
		scopeListString := ipInfo[findScopeIndex+7:]
		scopeList := strings.Split(scopeListString, " ")

		tmp.Scope = append(tmp.Scope, scopeList...)
		valid_lft := extInfo[findValidLftIndex+10 : findPreferredLftIndex-4]
		preferred_lft := extInfo[findPreferredLftIndex+14 : len(extInfo)-3]

		tmp.ValidTime, _ = strconv.Atoi(valid_lft)
		tmp.Preferred, _ = strconv.Atoi(preferred_lft)

		//fmt.Println(ipaddrStr, ipMark, scopeList, extInfo, valid_lft, "/", preferred_lft)

		result = append(result, tmp)
		//fmt.Println(i, ":", ipInfo, "-", extInfo)

	}

	//fmt.Println(outputCmd, "\n", string(resultBytes))
	//_ = resultBytes

	return result
}

func SelectIpV6TargetInfo(info []*Ipv6Info) *Ipv6Info {

	for _, ipv6Info := range info {
		for _, targetScpoe := range ipv6Info.Scope {
			if targetScpoe == "mngtmpaddr" {
				continue
			}
		}

		if ipv6Info.Mark == "128" {
			continue
		}

		return ipv6Info
	}

	return nil

}
