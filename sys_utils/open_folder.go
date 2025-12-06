package sys_utils

import (
	"os/exec"
)

// OpenFolder 打开指定文件夹（Windows Explorer）
func OpenFolder(path string) error {
	cmd := exec.Command("explorer", path)
	return cmd.Start()
}
