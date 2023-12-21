package control

// 配置文件

type DomainRecordInfo struct {
	RR         string `json:"RR,omitempty" xml:"RR,omitempty" yaml:"RR" mapstructure:"RR"`
	RecordId   string `json:"RecordId,omitempty" xml:"RecordId,omitempty" yaml:"RecordId" mapstructure:"RecordId"`
	Type       string `json:"Type,omitempty" xml:"Type,omitempty" yaml:"Type" mapstructure:"Type"`
	WatchType  string `json:"watch_type" mapstructure:"WatchType" yaml:"WatchType"`  // 观察类型 local -- 本机  esxi - 远程 esxi
	VMName     string `json:"vm_name" mapstructure:"VMName" yaml:"VMName,omitempty"` // esxi 的时候所属的 实例名
	LastVMAddr string `json:"-" mapstructure:"-" yaml:"-"`
	//Value    *string `json:"Value,omitempty" xml:"Value,omitempty" mapstructure:"Value"`
}

type EsxiConfig struct {
	Url      string `json:"url" mapstructure:"url"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
	Insecure bool   `json:"insecure" mapstructure:"insecure"`
}

//type Config struct {
//	InterfaceName string `json:"net_name" mapstructure:"net_name"`
//	LastIpV6Info  string `json:"-"`
//}

type LoggerConfig struct {
	LoggerPath  string `json:"logger_path"    yaml:"logger_path"    mapstructure:"logger_path"`
	LoggerLevel string `json:"logger_level"   yaml:"logger_level"   mapstructure:"logger_level"`
}

type AliyunConfig struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" xml:"accessKeyId,omitempty" mapstructure:"access_key_id"`
	AccessKeySecret string `json:"accessKeySecret,omitempty" xml:"accessKeySecret,omitempty" mapstructure:"access_key_secret"`
	Endpoint        string `json:"endpoint" mapstructure:"endpoint"`
}

type AppConfig struct {
	UpdateWaitTime int `json:"update_wait_time" mapstructure:"update_wait_time"`
	RetryTime      int `json:"retry_time" mapstructure:"retry_time"`
}

type AuthConfig struct {
	Token           string `json:"token"`
	TokenHeaderName string `json:"token_header_name"` //解析token 的时候的name 空则默认
}

type ConfigFile struct {
	Logger      LoggerConfig `json:"logger" mapstructure:"logger"`
	Esxi        EsxiConfig   `json:"esxi" mapstructure:"esxi"`
	Aliyun      AliyunConfig `json:"aliyun" mapstructure:"aliyun"`
	App         AppConfig    `json:"app" mapstructure:"app"`
	RecordsFile string       `json:"records_file" mapstructure:"records_file"`
}

type RecordsConfig struct {
	Records []*DomainRecordInfo `json:"records" mapstructure:"records"`
}
