package utils

import (
	"strings"

	"cron-shot/constants"

	"github.com/dlclark/regexp2"
)

// ResolveStorageFolder 使用给定正则/文本解析窗口标题，返回文件夹名
// 优先返回第一个捕获组，其次返回整体匹配；失败返回未知名称
func ResolveStorageFolder(windowTitle, storageRule string) string {
	rule := strings.TrimSpace(storageRule)
	if rule == "" {
		return SanitizeFolderName(windowTitle)
	}
	re, err := regexp2.Compile(rule, 0)
	if err != nil {
		return constants.TextUnknownName
	}
	m, err := re.FindStringMatch(windowTitle)
	if err != nil || m == nil {
		return constants.TextUnknownName
	}
	gps := m.Groups()
	if len(gps) > 1 {
		g := strings.TrimSpace(gps[1].String())
		if g != "" {
			return SanitizeFolderName(g)
		}
	}
	return SanitizeFolderName(m.String())
}
