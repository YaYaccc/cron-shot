package utils

import (
	"strings"
	"unicode"

	"cron-shot/constants"
)

// SanitizeProcessName 清理进程名，移除 .exe 后缀与非法字符
func SanitizeProcessName(name string) string {
	n := strings.TrimSpace(name)
	if i := strings.LastIndex(strings.ToLower(n), ".exe"); i != -1 && i+4 == len(n) {
		n = n[:i]
	}
	return SanitizeFolderName(n)
}

// SanitizeFolderName 清理文件夹名的非法字符与空白
func SanitizeFolderName(name string) string {
	s := strings.TrimSpace(name)
	invalid := []rune{'<', '>', ':', '\\', '/', '|', '?', '*', '"'}
	m := make(map[rune]struct{}, len(invalid))
	for _, r := range invalid {
		m[r] = struct{}{}
	}
	var b strings.Builder
	for _, r := range s {
		if _, bad := m[r]; bad {
			continue
		}
		if r == 0 || unicode.IsSpace(r) || r < 0x20 {
			continue
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return constants.TextUnknownName
	}
	return out
}
