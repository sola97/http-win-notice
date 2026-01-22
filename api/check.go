package api

import (
	"github.com/gin-gonic/gin"
	"http-win-notice/frp"
	"http-win-notice/model"
	"http-win-notice/utils/setting"
	"net/http"
	"os"
	"time"
)

type CheckItem struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type SelfCheckResponse struct {
	Ok     bool                 `json:"ok"`
	Time   string               `json:"time"`
	Checks map[string]CheckItem `json:"checks"`
}

func SelfCheckApi(c *gin.Context) {
	checks := make(map[string]CheckItem)
	overall := true
	setCheck := func(key string, ok bool, message string) {
		checks[key] = CheckItem{
			Ok:      ok,
			Message: message,
		}
		if !ok {
			overall = false
		}
	}

	if err := model.PingDb(); err != nil {
		setCheck("db", false, err.Error())
	} else {
		setCheck("db", true, "")
	}

	if info, err := os.Stat(setting.ConfigPath); err != nil {
		setCheck("config", false, err.Error())
	} else if info.IsDir() {
		setCheck("config", false, "config path is a directory")
	} else {
		setCheck("config", true, "")
	}

	if info, err := os.Stat(setting.ImagesDir); err != nil {
		setCheck("images_dir", false, err.Error())
	} else if !info.IsDir() {
		setCheck("images_dir", false, "images path is not a directory")
	} else {
		setCheck("images_dir", true, "")
	}

	frpStatus := frp.GetStatus()
	if frpStatus.Enabled {
		if frpStatus.Running {
			setCheck("frp", true, "")
		} else if frpStatus.LastError != "" {
			setCheck("frp", false, frpStatus.LastError)
		} else {
			setCheck("frp", false, "frpc not running")
		}
	} else {
		setCheck("frp", true, "disabled")
	}

	c.JSON(http.StatusOK, SelfCheckResponse{
		Ok:     overall,
		Time:   time.Now().Format("2006-01-02 15:04:05"),
		Checks: checks,
	})
}
