package win

import (
	"cron-shot/config"
	"cron-shot/constants"
	"cron-shot/sys_utils"
	"os"
	"strings"
)

// InitAutostartRegistration 根据配置同步应用的开机自启注册状态
// - 开启时：若未注册或注册路径与当前可执行文件不一致，则重新写入注册表
// - 关闭时：若已注册则移除注册表项
func InitAutostartRegistration() {
	app := constants.TextAppTitle
	if config.GetAutostartEnabled() {
		exe, _ := os.Executable()
		// 查询是否已注册以及注册值（可执行路径）
		ok, v, err := sys_utils.IsAutoStartRegistered(app)
		if err == nil {
			// 未注册或路径不一致时，写入当前 exe 路径
			if !ok || strings.TrimSpace(v) != exe {
				_ = sys_utils.EnableAutoStart(app, exe)
			}
		}
	} else {
		// 自启关闭时，如已注册则删除
		ok, _, err := sys_utils.IsAutoStartRegistered(app)
		if err == nil && ok {
			_ = sys_utils.DisableAutoStart(app)
		}
	}
}
