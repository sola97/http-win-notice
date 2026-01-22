package setting

import (
	"gopkg.in/yaml.v3"
	"http-win-notice/utils"
	"http-win-notice/utils/b64"
	"http-win-notice/utils/constant"
	"http-win-notice/utils/log"
	"os"
	"path"
	"strconv"
)

var ConfigPath string
var Config ConfigType

var DbPath string
var LogoPath string
var ImagesDir string

var Version string = "v0.0.0"

type ConfigType struct {
	Debug   int         `yaml:"debug" json:"debug"`
	Port    int         `yaml:"port" json:"port"`
	Toast   ToastType   `yaml:"toast" json:"toast"`
	Frp     FrpType     `yaml:"frp" json:"frp"`
	Forward ForwardType `yaml:"forward" json:"forward"`
}

type ToastType struct {
	Icons   []string          `yaml:"icons" json:"icons"`
	IconMap map[string]string `yaml:"icon_map" json:"icon_map"`
}

type FrpType struct {
	Enable     int          `yaml:"enable" json:"enable"`
	ServerAddr string       `yaml:"server_addr" json:"server_addr"`
	ServerPort int          `yaml:"server_port" json:"server_port"`
	Token      string       `yaml:"token" json:"token"`
	Proxy      FrpProxyType `yaml:"proxy" json:"proxy"`
}

type FrpProxyType struct {
	Name       string `yaml:"name" json:"name"`
	Type       string `yaml:"type" json:"type"`
	LocalIP    string `yaml:"local_ip" json:"local_ip"`
	LocalPort  int    `yaml:"local_port" json:"local_port"`
	RemotePort int    `yaml:"remote_port" json:"remote_port"`
}

type ForwardType struct {
	Enable     int    `yaml:"enable" json:"enable"`
	ServerAddr string `yaml:"server_addr" json:"server_addr"`
	ServerPort int    `yaml:"server_port" json:"server_port"`
	Token      string `yaml:"token" json:"token"`
}

func init() {
	ConfigPath = path.Join(utils.RootDir, constant.ConfigFile)
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		initConfigFile(ConfigPath)
	}
	InitConfig()
	DbPath = path.Join(utils.RootDir, constant.DbFile)

	ImagesDir = path.Join(utils.RootDir, "images")
	if _, err := os.Stat(ImagesDir); os.IsNotExist(err) {
		err = os.Mkdir(ImagesDir, os.ModeDir)
		if err != nil {
			log.Fatalln(err)
		}
	}

	LogoPath = path.Join(ImagesDir, "logo.png")
	if _, err := os.Stat(LogoPath); os.IsNotExist(err) {
		b64.WriteImgFileByBase64(constant.LogoImgBase64, LogoPath)
	}

}

func initConfigFile(ConfigPath string) {
	//默认配置
	Config.Debug = 0
	Config.Port = 19000
	Frp := FrpType{
		Enable:     0,
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
		Token:      "httpwinnotice123456",
		Proxy: FrpProxyType{
			Name:       "http-win-notice",
			Type:       "tcp",
			LocalIP:    "127.0.0.1",
			LocalPort:  Config.Port,
			RemotePort: 19001,
		},
	}
	Config.Frp = Frp

	icons := []string{"logo.png"}
	Toast := ToastType{
		Icons:   icons,
		IconMap: map[string]string{"default": "logo.png"},
	}
	Config.Toast = Toast

	//默认转发配置
	Forward := ForwardType{
		Enable:     0,
		ServerAddr: "127.0.0.1",
		ServerPort: 9919,
		Token:      "httpwinnotice123456",
	}
	Config.Forward = Forward

	updatedData, err := yaml.Marshal(Config)
	if err != nil {
		log.Fatalln("create config error", err)
	}
	log.Info("create config path", ConfigPath)
	err = os.WriteFile(ConfigPath, updatedData, 0644)
}

func InitConfig() {
	log.Info("start init config")
	Config = ConfigType{}
	dataBytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		log.Fatalln("读取配置文件失败")
	}

	err = yaml.Unmarshal(dataBytes, &Config)
	if err != nil {
		log.Fatalln("解析config失败")
	}

	if Config.Debug == 1 {
		log.SetLogLevel("debug")
	}
	log.Debug("config=", Config)
}

func PortStr() string {
	return strconv.Itoa(Config.Port)
}

func HomeUrl() string {
	return constant.BaseUrl + ":" + PortStr()
}
