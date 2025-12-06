package sys_utils

import (
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// GetProcessNames 获取当前运行的所有进程名称，并进行去重和排序
func GetProcessNames() ([]string, error) {
	// 获取所有进程
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	// 使用 map 进行去重
	nameSet := make(map[string]struct{})
	for _, p := range procs {
		name, err := p.Name()
		if err == nil && name != "" {
			nameSet[name] = struct{}{}
		}
	}

	// 将 map 转换为切片
	var processNames []string
	for name := range nameSet {
		processNames = append(processNames, name)
	}

	// 排序（忽略大小写）
	sort.Slice(processNames, func(i, j int) bool {
		return strings.ToLower(processNames[i]) < strings.ToLower(processNames[j])
	})

	return processNames, nil
}
