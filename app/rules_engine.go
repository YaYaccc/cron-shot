package app

import (
	"cron-shot/config"
	"cron-shot/utils"
	"regexp"
)

// MatchRule 在规则列表中查找匹配窗口标题的规则
// 优先文本等价匹配，其次正则匹配；返回匹配到的规则与是否命中
func MatchRule(title string, rules []config.AppRule) (*config.AppRule, bool) {
	// 一次线性扫描做等价匹配（更快且避免不必要的正则编译）
	for i := range rules {
		r := &rules[i]
		if !r.Enabled {
			continue
		}
		if r.Pattern == title {
			return r, true
		}
	}
	// 兜底：逐条编译正则并尝试匹配
	for i := range rules {
		r := &rules[i]
		if !r.Enabled {
			continue
		}
		re, err := regexp.Compile(r.Pattern)
		if err == nil && re.MatchString(title) {
			return r, true
		}
	}
	return nil, false
}

// ResolveFolder 根据规则解析存储文件夹和固定前缀
// 当 rule 为 nil 时返回标题作为子文件夹；否则使用 StorageRule 解析
func ResolveFolder(title string, rule *config.AppRule) (string, string) {
	if rule == nil {
		return title, ""
	}
	// 使用正则捕获组/整体匹配解析出文件夹名，并进行安全化处理
	folder := utils.ResolveStorageFolder(title, rule.StorageRule)
	return folder, rule.FixedFolder
}
