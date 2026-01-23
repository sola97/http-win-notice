package notice

import (
	"http-win-notice/utils/log"
	"http-win-notice/utils/setting"
	"regexp"
	"strings"
)

func ShouldFilter(title, msg, remark string) (bool, string) {
	cfg := setting.Config.Filter
	if cfg.Enable != 1 {
		return false, ""
	}

	allText := strings.Join([]string{title, msg, remark}, "\n")

	if len(cfg.Whitelist) > 0 {
		if ok, _ := matchAny(cfg.Whitelist, title, msg, remark, allText); !ok {
			return true, "not in whitelist"
		}
	}

	if ok, detail := matchAny(cfg.Blacklist, title, msg, remark, allText); ok {
		return true, "blacklist match: " + detail
	}

	return false, ""
}

func matchAny(rules []setting.FilterRule, title, msg, remark, allText string) (bool, string) {
	for _, rule := range rules {
		if rule.Enabled != 1 || strings.TrimSpace(rule.Pattern) == "" {
			continue
		}
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			log.Println("filter regex invalid:", rule.Pattern, err)
			continue
		}
		if re.MatchString(pickFieldText(rule.Field, title, msg, remark, allText)) {
			return true, ruleDetail(rule)
		}
	}
	return false, ""
}

func pickFieldText(field, title, msg, remark, allText string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "title":
		return title
	case "msg":
		return msg
	case "remark":
		return remark
	default:
		return allText
	}
}

func ruleDetail(rule setting.FilterRule) string {
	field := strings.TrimSpace(rule.Field)
	pattern := strings.TrimSpace(rule.Pattern)
	if field == "" {
		field = "all"
	}
	if pattern == "" {
		pattern = "-"
	}
	return "field=" + field + " pattern=" + pattern
}
