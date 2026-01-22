package notice

import (
	"github.com/go-toast/toast"
	"http-win-notice/utils/constant"
	"http-win-notice/utils/setting"
	"os"
	"path"
	"strconv"
	"strings"
)

type Action = toast.Action

func notificationSetAttr(n *toast.Notification, duration string, audio string) {
	switch duration {
	case "long":
		n.Duration = toast.Long
	case "short":
		n.Duration = toast.Short
	default:
		n.Duration = toast.Short
	}

	switch audio {
	case "default":
		n.Audio = toast.Default
	case "im":
		n.Audio = toast.IM
	case "mail":
		n.Audio = toast.Mail
	case "reminder":
		n.Audio = toast.Reminder
	case "sms":
		n.Audio = toast.SMS
	case "loopingalarm":
		n.Audio = toast.LoopingAlarm
	case "loopingalarm2":
		n.Audio = toast.LoopingAlarm2
	case "loopingalarm3":
		n.Audio = toast.LoopingAlarm3
	case "loopingalarm4":
		n.Audio = toast.LoopingAlarm4
	case "loopingalarm5":
		n.Audio = toast.LoopingAlarm5
	case "loopingalarm6":
		n.Audio = toast.LoopingAlarm6
	case "loopingalarm7":
		n.Audio = toast.LoopingAlarm7
	case "loopingalarm8":
		n.Audio = toast.LoopingAlarm8
	case "loopingalarm9":
		n.Audio = toast.LoopingAlarm9
	case "loopingalarm10":
		n.Audio = toast.LoopingAlarm10
	case "loopingcall":
		n.Audio = toast.LoopingCall
	case "loopingcall2":
		n.Audio = toast.LoopingCall2
	case "loopingcall3":
		n.Audio = toast.LoopingCall3
	case "loopingcall4":
		n.Audio = toast.LoopingCall4
	case "loopingcall5":
		n.Audio = toast.LoopingCall5
	case "loopingcall6":
		n.Audio = toast.LoopingCall6
	case "loopingcall7":
		n.Audio = toast.LoopingCall7
	case "loopingcall8":
		n.Audio = toast.LoopingCall8
	case "loopingcall9":
		n.Audio = toast.LoopingCall9
	case "loopingcall10":
		n.Audio = toast.LoopingCall10
	case "silent":
		n.Audio = toast.Silent
	default:
		n.Audio = toast.Default
	}

}

func Notice(msg, title, icon, duration, audio string, actions []toast.Action) error {
	iconPath := resolveIconPath(icon)
	for i := range actions {
		if actions[i].Arguments == "$WEB_PAGE" {
			actions[i].Arguments = setting.HomeUrl()
		}
	}
	if len(actions) == 0 {
		actions = []Action{
			{"protocol", "查看", setting.HomeUrl()},
		}
	}
	notification := toast.Notification{
		AppID:   constant.AppID,
		Title:   title,
		Message: msg,
		Icon:    iconPath,
		Actions: actions,
	}
	notificationSetAttr(&notification, duration, audio)
	//fmt.Println("n=", notification)
	err := notification.Push()
	return err
	//if err != nil {
	//	panic(err)
	//}
}

func resolveIconPath(icon string) string {
	file := resolveIconFile(icon)
	return path.Join(setting.ImagesDir, file)
}

func resolveIconFile(icon string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return pickDefaultIconFile()
	}
	if idx, err := strconv.Atoi(icon); err == nil {
		if idx >= 0 && idx < len(setting.Config.Toast.Icons) {
			file := normalizeIconName(setting.Config.Toast.Icons[idx])
			if iconFileExists(file) {
				return file
			}
		}
		return pickDefaultIconFile()
	}
	if file, ok := setting.Config.Toast.IconMap[icon]; ok {
		if resolved, ok := resolveIconFromName(file); ok {
			return resolved
		}
	}
	if resolved, ok := resolveIconFromName(icon); ok {
		return resolved
	}
	return pickDefaultIconFile()
}

func pickDefaultIconFile() string {
	for _, candidate := range setting.Config.Toast.Icons {
		if resolved, ok := resolveIconFromName(candidate); ok {
			return resolved
		}
	}
	if resolved, ok := resolveIconFromName("logo.png"); ok {
		return resolved
	}
	if len(setting.Config.Toast.Icons) > 0 {
		file := normalizeIconName(setting.Config.Toast.Icons[0])
		if file != "" {
			return file
		}
	}
	return "logo.png"
}

func normalizeIconName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	if name == "." || name == "/" {
		return ""
	}
	return name
}

func iconFileExists(name string) bool {
	if name == "" {
		return false
	}
	_, err := os.Stat(path.Join(setting.ImagesDir, name))
	return err == nil
}

func isAllowedIconExt(name string) bool {
	ext := strings.ToLower(path.Ext(name))
	for _, allowed := range allowedIconExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

var allowedIconExts = []string{".png", ".jpg", ".jpeg", ".ico"}

func resolveIconFromName(name string) (string, bool) {
	name = normalizeIconName(name)
	if name == "" {
		return "", false
	}
	if isAllowedIconExt(name) {
		if iconFileExists(name) {
			return name, true
		}
		return "", false
	}
	for _, ext := range allowedIconExts {
		candidate := name + ext
		if iconFileExists(candidate) {
			return candidate, true
		}
	}
	return "", false
}
