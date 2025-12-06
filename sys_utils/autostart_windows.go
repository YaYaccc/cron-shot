package sys_utils

import (
	"golang.org/x/sys/windows/registry"
)

// EnableAutoStart 在注册表 Run 项中写入启动项以启用开机自启动
func EnableAutoStart(appName, exePath string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(appName, exePath)
}

// DisableAutoStart 从注册表 Run 项中移除启动项以禁用开机自启动
func DisableAutoStart(appName string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	// ignore error when value not exist
	_ = k.DeleteValue(appName)
	return nil
}
