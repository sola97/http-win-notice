package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"http-win-notice/model"
	"http-win-notice/utils/constant"
	"http-win-notice/utils/notice"
	"net/http"
	"strconv"
	"strings"
)

type Button struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type ToastRequest struct {
	Title    string   `json:"title"`
	Msg      string   `json:"msg"`
	Remark   string   `json:"remark"`
	Icon     string   `json:"icon"`
	Duration string   `json:"duration"`
	Audio    string   `json:"audio"`
	Button   []Button `json:"button"`
}

func ToastApi(c *gin.Context) {
	req := ToastRequest{
		Msg:      c.DefaultQuery("msg", constant.DefaultMsg),
		Title:    c.DefaultQuery("title", constant.DefaultTitle),
		Icon:     c.DefaultQuery("icon", constant.DefaultIcon),
		Duration: c.DefaultQuery("duration", constant.DefaultDuration),
		Audio:    c.DefaultQuery("audio", constant.DefaultAudio),
		Remark:   c.Query("remark"),
	}
	button := c.DefaultQuery("button", "")
	if button != "" {
		_ = json.Unmarshal([]byte(button), &req.Button)
	}
	normalizeToastRequest(&req)
	handleToastRequest(c, req)
}

func ToastPostApi(c *gin.Context) {
	var req ToastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request body",
		})
		return
	}
	normalizeToastRequest(&req)
	handleToastRequest(c, req)
}

func HomePage(c *gin.Context) {
	page := c.Query("page")
	pageInt64, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		pageInt64 = 1
	}
	pageInt := int(pageInt64)
	sizeInt := constant.DefaultWebPageSize
	msgList := model.SearchMsg(pageInt, sizeInt)
	formatMsgList := make([]gin.H, len(msgList))
	total := model.TotalMsg()
	for i, msg := range msgList {
		var statusShow string
		switch msg.Status {
		case 1:
			statusShow = "已提醒"
		case 9:
			statusShow = "提醒失败"
		default:
			statusShow = "等待"
		}
		formatMsg := gin.H{
			"id":     msg.ID,
			"index":  (pageInt-1)*10 + i + 1,
			"title":  msg.Title,
			"msg":    msg.Msg,
			"remark": msg.Remark,
			"time":   msg.CreatedAt.Format("2006-01-02 15:04:05"),
			"status": statusShow,
		}
		formatMsgList[i] = formatMsg
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"msgList":     formatMsgList,
		"total":       total,
		"currentPage": pageInt,
		"pageSize":    sizeInt,
	})
}

func ClearLogsApi(c *gin.Context) {
	deleted, err := model.ClearMsgAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to clear logs",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"deleted": deleted,
	})
}

func normalizeToastRequest(req *ToastRequest) {
	req.Msg = strings.TrimSpace(req.Msg)
	if req.Msg == "" {
		req.Msg = constant.DefaultMsg
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		req.Title = constant.DefaultTitle
	}
	req.Icon = strings.TrimSpace(req.Icon)
	if req.Icon == "" {
		req.Icon = constant.DefaultIcon
	}
	req.Duration = strings.TrimSpace(req.Duration)
	if req.Duration == "" {
		req.Duration = constant.DefaultDuration
	}
	req.Audio = strings.TrimSpace(req.Audio)
	if req.Audio == "" {
		req.Audio = constant.DefaultAudio
	}
	req.Remark = strings.TrimSpace(req.Remark)
}

func handleToastRequest(c *gin.Context, req ToastRequest) {
	actions := actionsFromButtons(req.Button)
	data := model.Msg{
		Title:  req.Title,
		Msg:    req.Msg,
		Remark: req.Remark,
	}
	res := model.CreateMsg(&data)
	if !res {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "fail, db error",
		})
		return
	}
	go func() {
		err := notice.Notice(req.Msg, req.Title, req.Icon, req.Duration, req.Audio, actions)
		if err != nil {
			model.UpdateMsgStatus(data.ID, 9)
		} else {
			model.UpdateMsgStatus(data.ID, 1)
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func actionsFromButtons(buttons []Button) []notice.Action {
	if len(buttons) == 0 {
		return nil
	}
	actions := make([]notice.Action, 0, len(buttons))
	for _, bt := range buttons {
		label := strings.TrimSpace(bt.Name)
		url := strings.TrimSpace(bt.Url)
		if label == "" || url == "" {
			continue
		}
		actions = append(actions, notice.Action{
			Type:      "protocol",
			Label:     label,
			Arguments: url,
		})
	}
	return actions
}
