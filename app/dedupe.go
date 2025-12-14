package app

import (
	"cron-shot/config"
	"cron-shot/utils"
	"image"
	"path/filepath"
	"strings"
)

// ShouldSkipDueToDedupe 根据阈值判断是否跳过保存（去重）
// - 当阈值=100时执行像素级全等比较
// - 否则使用 AHash16x16 + 汉明距离计算相似度
func ShouldSkipDueToDedupe(img *image.RGBA, storageRoot, processName, fixed, folder string) bool {
	// 去重开关关闭则直接保存
	if !config.GetDedupeEnabled() {
		return false
	}
	// 构造目标目录：process/fixed/folder 或 process/folder
	proc := utils.SanitizeProcessName(processName)
	sub := utils.SanitizeFolderName(folder)
	var dir string
	if strings.TrimSpace(fixed) != "" {
		fix := utils.SanitizeFolderName(fixed)
		dir = filepath.Join(storageRoot, proc, fix, sub)
	} else {
		dir = filepath.Join(storageRoot, proc, sub)
	}
	// 读取最近一张图片；无历史则不跳过
	prevImg, _ := utils.LatestPNGImage(dir)
	if prevImg == nil {
		return false
	}
	th := config.GetDedupeThreshold()
	if th >= 100 {
		// 阈值满分：执行像素级比较
		return utils.ImagesEqualExact(img, prevImg)
	}
	// 计算感知哈希并转换为相似度百分比
	prevHash := utils.AHashFromImage(prevImg)
	currHash := utils.AHash16x16(img)
	dist := utils.Hamming256(prevHash, currHash)
	sim := 1.0 - float64(dist)/256.0
	return sim*100.0 >= float64(th)
}
