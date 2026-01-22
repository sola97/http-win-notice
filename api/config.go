package api

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"http-win-notice/utils/setting"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

type ConfigView struct {
	Debug int               `json:"debug"`
	Port  int               `json:"port"`
	Toast setting.ToastType `json:"toast"`
	Frp   setting.FrpType   `json:"frp"`
}

type ConfigUpdate = ConfigView

func ConfigPage(c *gin.Context) {
	c.HTML(http.StatusOK, "config.html", gin.H{})
}

func GetConfigApi(c *gin.Context) {
	config := ConfigView{
		Debug: setting.Config.Debug,
		Port:  setting.Config.Port,
		Toast: setting.Config.Toast,
		Frp:   setting.Config.Frp,
	}
	c.JSON(http.StatusOK, gin.H{
		"config":            config,
		"available_icons":   listIconFiles(),
		"images_url_prefix": "/images/",
	})
}

func UpdateConfigApi(c *gin.Context) {
	var req ConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}

	if req.Port <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "port must be greater than 0",
		})
		return
	}

	req.Debug = normalizeToggle(req.Debug)
	req.Frp.Enable = normalizeToggle(req.Frp.Enable)
	req.Toast.Icons = normalizeIcons(req.Toast.Icons)
	req.Toast.IconMap = normalizeIconMap(req.Toast.IconMap)
	req.Toast.Icons = mergeIconMapValues(req.Toast.Icons, req.Toast.IconMap)
	normalizeFrp(&req.Frp, req.Port)

	if req.Frp.Enable == 1 {
		if req.Frp.ServerAddr == "" || req.Frp.ServerPort <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "frp server address/port is required",
			})
			return
		}
		if req.Frp.Proxy.Name == "" || req.Frp.Proxy.LocalPort <= 0 || req.Frp.Proxy.RemotePort <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "frp proxy name/local_port/remote_port is required",
			})
			return
		}
	}

	setting.Config.Debug = req.Debug
	setting.Config.Port = req.Port
	setting.Config.Toast = req.Toast
	setting.Config.Frp = req.Frp

	if err := writeConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to write config",
		})
		return
	}
	setting.InitConfig()

	c.JSON(http.StatusOK, gin.H{
		"message":           "success",
		"restart_required":  true,
		"images_url_prefix": "/images/",
	})
}

func UploadIconApi(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "file is required",
		})
		return
	}

	filename := sanitizeFilename(file.Filename)
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid filename",
		})
		return
	}
	if !isAllowedIconExt(filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "only .png, .jpg, .jpeg, .ico are allowed",
		})
		return
	}

	dst := path.Join(setting.ImagesDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
	})
}

func writeConfig() error {
	updatedData, err := yaml.Marshal(setting.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(setting.ConfigPath, updatedData, 0644)
}

func normalizeToggle(value int) int {
	if value != 0 {
		return 1
	}
	return 0
}

func normalizeIcons(icons []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, icon := range icons {
		name := sanitizeFilename(strings.TrimSpace(icon))
		if name == "" || !isAllowedIconExt(name) || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func normalizeIconMap(iconMap map[string]string) map[string]string {
	out := make(map[string]string)
	for key, value := range iconMap {
		alias := strings.TrimSpace(key)
		file := sanitizeFilename(strings.TrimSpace(value))
		if alias == "" || file == "" || !isAllowedIconExt(file) {
			continue
		}
		out[alias] = file
	}
	return out
}

func mergeIconMapValues(icons []string, iconMap map[string]string) []string {
	if len(iconMap) == 0 {
		return icons
	}
	seen := make(map[string]bool, len(icons))
	for _, icon := range icons {
		seen[icon] = true
	}
	for _, file := range iconMap {
		if !seen[file] {
			icons = append(icons, file)
			seen[file] = true
		}
	}
	return icons
}

func normalizeFrp(frp *setting.FrpType, defaultLocalPort int) {
	if frp.ServerAddr == "" {
		frp.ServerAddr = "127.0.0.1"
	}
	if frp.ServerPort <= 0 {
		frp.ServerPort = 7000
	}
	if frp.Proxy.Name == "" {
		frp.Proxy.Name = "http-win-notice"
	}
	if frp.Proxy.Type == "" {
		frp.Proxy.Type = "tcp"
	}
	if frp.Proxy.LocalIP == "" {
		frp.Proxy.LocalIP = "127.0.0.1"
	}
	if frp.Proxy.LocalPort <= 0 {
		frp.Proxy.LocalPort = defaultLocalPort
	}
	if frp.Proxy.RemotePort <= 0 {
		frp.Proxy.RemotePort = 19001
	}
}

func listIconFiles() []string {
	entries, err := os.ReadDir(setting.ImagesDir)
	if err != nil {
		return []string{}
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if isAllowedIconExt(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	if name == "." || name == "/" {
		return ""
	}
	return name
}

func isAllowedIconExt(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".ico":
		return true
	default:
		return false
	}
}
