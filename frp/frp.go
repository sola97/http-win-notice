package frp

import (
	"fmt"
	"http-win-notice/utils"
	"http-win-notice/utils/log"
	"http-win-notice/utils/setting"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const FrpConfigTemplate = `[common]
server_addr = ${ServerAddr}
server_port = ${ServerPort}
token = ${Token}

[${ProxyName}]
type = ${ProxyType}
local_ip = ${LocalIP}
local_port = ${LocalPort}
remote_port = ${RemotePort}`

type RuntimeStatus struct {
	Enabled    bool   `json:"enabled"`
	Running    bool   `json:"running"`
	PID        int    `json:"pid"`
	StartTime  string `json:"start_time"`
	LastError  string `json:"last_error"`
	FrpcPath   string `json:"frpc_path"`
	ConfigPath string `json:"config_path"`
	LogPath    string `json:"log_path"`
}

var (
	statusMu sync.Mutex
	status   = baseStatus()
)

func baseStatus() RuntimeStatus {
	return RuntimeStatus{
		Enabled:    setting.Config.Frp.Enable == 1,
		Running:    false,
		PID:        0,
		StartTime:  "",
		LastError:  "",
		FrpcPath:   path.Join(utils.RootDir, "frpc.exe"),
		ConfigPath: path.Join(utils.RootDir, "frpc.ini"),
		LogPath:    path.Join(utils.RootDir, "frpc.log"),
	}
}

func updateStatus(update func(*RuntimeStatus)) {
	statusMu.Lock()
	update(&status)
	statusMu.Unlock()
}

func GetStatus() RuntimeStatus {
	statusMu.Lock()
	current := status
	statusMu.Unlock()
	current.Enabled = setting.Config.Frp.Enable == 1
	if current.FrpcPath == "" || current.ConfigPath == "" || current.LogPath == "" {
		base := baseStatus()
		if current.FrpcPath == "" {
			current.FrpcPath = base.FrpcPath
		}
		if current.ConfigPath == "" {
			current.ConfigPath = base.ConfigPath
		}
		if current.LogPath == "" {
			current.LogPath = base.LogPath
		}
	}
	return current
}

func RestartFrp() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("frp restart is only supported on windows")
	}
	current := GetStatus()
	err := stopFrpProcess(current)
	updateStatus(func(s *RuntimeStatus) {
		s.Running = false
		s.PID = 0
		if err != nil {
			s.LastError = err.Error()
		} else {
			s.LastError = ""
		}
	})
	if err != nil {
		log.Println("frp stop failed:", err)
		return err
	}
	if setting.Config.Frp.Enable != 1 {
		return fmt.Errorf("frp is disabled")
	}
	StartFrp()
	return nil
}

func stopFrpProcess(status RuntimeStatus) error {
	if status.PID > 0 {
		return runTaskkill("/F", "/T", "/PID", strconv.Itoa(status.PID))
	}
	return runTaskkill("/F", "/T", "/IM", "frpc.exe")
}

func runTaskkill(args ...string) error {
	cmd := exec.Command("taskkill", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "not found") || strings.Contains(lower, "no instance") || strings.Contains(lower, "not running") {
			return nil
		}
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("taskkill failed: %s", msg)
	}
	return nil
}

func initFrpConfig() {
	proxy := setting.Config.Frp.Proxy
	if proxy.Name == "" {
		proxy.Name = "http-win-notice"
	}
	if proxy.Type == "" {
		proxy.Type = "tcp"
	}
	if proxy.LocalIP == "" {
		proxy.LocalIP = "127.0.0.1"
	}
	if proxy.LocalPort == 0 {
		proxy.LocalPort = setting.Config.Port
	}
	if proxy.RemotePort == 0 {
		proxy.RemotePort = 19001
	}
	var config = map[string]string{
		"ServerAddr": setting.Config.Frp.ServerAddr,
		"ServerPort": strconv.Itoa(setting.Config.Frp.ServerPort),
		"Token":      setting.Config.Frp.Token,
		"ProxyName":  proxy.Name,
		"ProxyType":  proxy.Type,
		"LocalIP":    proxy.LocalIP,
		"LocalPort":  strconv.Itoa(proxy.LocalPort),
		"RemotePort": strconv.Itoa(proxy.RemotePort),
	}
	s := os.Expand(FrpConfigTemplate, func(k string) string {
		return config[k]
	})
	configPath := path.Join(utils.RootDir, "frpc.ini")
	err := os.WriteFile(configPath, []byte(s), 0644)
	if err != nil {
		log.Fatalln("frp配置生成失败", err)
	}
}

func StartFrp() {
	go func() {
		base := baseStatus()
		updateStatus(func(s *RuntimeStatus) {
			*s = base
		})
		if setting.Config.Frp.Enable != 1 {
			return
		}
		initFrpConfig()
		startTime := time.Now().Format(time.DateTime)
		updateStatus(func(s *RuntimeStatus) {
			s.StartTime = startTime
		})
		frpExePath := base.FrpcPath
		if _, err := os.Stat(frpExePath); os.IsNotExist(err) {
			log.Println("启动frpc失败: " + frpExePath + "不存在")
			updateStatus(func(s *RuntimeStatus) {
				s.LastError = "frpc.exe not found"
			})
			return
		}
		configPath := base.ConfigPath
		logFile, err := os.OpenFile(base.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Println("frp日志文件打开失败", err)
			updateStatus(func(s *RuntimeStatus) {
				s.LastError = err.Error()
			})
			return
		}
		cmd := exec.Command(frpExePath, "-c", configPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		log.Println("start launch frpc")
		err = cmd.Start()
		if err != nil {
			log.Println("frp启动失败", err)
			updateStatus(func(s *RuntimeStatus) {
				s.LastError = err.Error()
				s.Running = false
			})
			_ = logFile.Close()
			return
		}
		updateStatus(func(s *RuntimeStatus) {
			s.Running = true
			if cmd.Process != nil {
				s.PID = cmd.Process.Pid
			}
		})
		_ = logFile.Close()
		go func() {
			waitErr := cmd.Wait()
			updateStatus(func(s *RuntimeStatus) {
				s.Running = false
				if waitErr != nil {
					s.LastError = waitErr.Error()
				}
			})
		}()
	}()

}
