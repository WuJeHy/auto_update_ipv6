package tools

import (
	"context"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"go.uber.org/zap"
	"net/netip"
	"net/url"
)

// 读取 数据

func GetAllEsxiAddrs(logger *zap.Logger, Url string,
	Username string,
	Password string,
	Insecure bool) (info map[string]string, err error) {
	//

	info = make(map[string]string, 8)

	// 可以读取配置

	// 实例化客户端
	ctxBg := context.Background()
	esxiClient, err := NewClient(ctxBg, Url, Username, Password, Insecure)
	if err != nil {
		return
	}

	// 读取配置信息

	readConfigFunc := func(ctx context.Context, c *vim25.Client) error {
		m := view.NewManager(c)

		v, errCreateContainerView := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
		if errCreateContainerView != nil {
			return errCreateContainerView
		}

		defer v.Destroy(ctx)

		var vms []mo.VirtualMachine
		errRetrieve := v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "guest"}, &vms)
		if errRetrieve != nil {
			return errRetrieve
		}

		for _, vm := range vms {
			//fmt.Printf("%s: %s\n", vm.Summary.Config.Name, vm.Summary.Config.GuestFullName)

			vmName := vm.Summary.Config.Name

			for _, netInfo := range vm.Guest.Net {
				for _, addrStr := range netInfo.IpAddress {
					//fmt.Println(i, index, addr)
					// 判断 ip 是否符合要求
					// todo 前5个字符是 2409: 的是移动的

					ipAddrInfo, errParseAddr := netip.ParseAddr(addrStr)
					if errParseAddr != nil {
						continue
					}
					// 判断是不是 v6

					if !ipAddrInfo.Is6() {
						continue
					}

					// v6 的ip

					// 判断开头

					targetIpv6String := ipAddrInfo.String()

					headerSub := targetIpv6String[:5]

					if headerSub == "2409:" {
						// 我们的目标 ipv6
						info[vmName] = targetIpv6String
					}

				}
			}
		}
		return nil
	}

	err = readConfigFunc(ctxBg, esxiClient)
	if err != nil {
		return
	}

	return

}

//

// NewClient creates a vim25.Client for use in the examples
func NewClient(ctx context.Context, url, username, password string, insecureFlag bool) (*vim25.Client, error) {
	// Parse URL from string
	u, err := soap.ParseURL(url)
	if err != nil {
		return nil, err
	}

	// Override username and/or password as required
	processOverride(u, username, password)

	// Share govc's session cache
	s := &cache.Session{
		URL:      u,
		Insecure: insecureFlag,
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func processOverride(u *url.URL, envUsername, envPassword string) {

	// Override username if provided
	if envUsername != "" {
		var password string
		var ok bool

		if u.User != nil {
			password, ok = u.User.Password()
		}

		if ok {
			u.User = url.UserPassword(envUsername, password)
		} else {
			u.User = url.User(envUsername)
		}
	}

	// Override password if provided
	if envPassword != "" {
		var username string

		if u.User != nil {
			username = u.User.Username()
		}

		u.User = url.UserPassword(username, envPassword)
	}
}
