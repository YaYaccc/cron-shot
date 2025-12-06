package sys_utils

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// GetProcessWindows 获取指定进程打开的所有可见窗口标题
// processName: 进程名称（例如 "chrome.exe" 或 "chrome"）
func GetProcessWindows(processName string) ([]string, error) {
	var windows []string

	// 确保进程名以 .exe 结尾以便匹配（如果输入不带 .exe）
	targetProcessName := processName
	if !strings.HasSuffix(strings.ToLower(targetProcessName), ".exe") {
		targetProcessName += ".exe"
	}

	// 使用 uintptr 替代 win.LPARAM 以避免 undefined 错误
	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		// 获取窗口所属的进程ID
		var processId uint32
		win.GetWindowThreadProcessId(hwnd, &processId)

		// 获取进程句柄
		const PROCESS_QUERY_INFORMATION = 0x0400
		const PROCESS_VM_READ = 0x0010

		hProcess, err := syscall.OpenProcess(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, false, processId)
		if err == nil && hProcess != 0 {
			defer syscall.CloseHandle(hProcess)

			// 获取进程名
			var name [win.MAX_PATH]uint16
			len := uint32(len(name))
			ret, _, _ := syscall.NewLazyDLL("psapi.dll").NewProc("GetModuleBaseNameW").Call(
				uintptr(hProcess),
				0,
				uintptr(unsafe.Pointer(&name[0])),
				uintptr(len),
			)

			if ret != 0 {
				pName := syscall.UTF16ToString(name[:])
				// 忽略大小写比较
				if strings.EqualFold(pName, targetProcessName) {
					// 获取窗口标题
					titleLen := getWindowTextLength(hwnd)
					if titleLen > 0 {
						buf := make([]uint16, titleLen+1)
						getWindowText(hwnd, &buf[0], int32(titleLen+1))
						title := syscall.UTF16ToString(buf)
						// 只要有标题就认为是有效窗口. 因为有些窗口可能是最小化状态或者被其他窗口遮挡， 。
						// 为了更宽松的匹配，我们只检查标题不为空。
						if title != "" && win.IsWindowVisible(hwnd) {
							windows = append(windows, title)
						}
					}
				}
			}
		}

		return 1 // 继续枚举
	})

	// 直接使用 syscall 调用 EnumWindows
	enumWindows(cb, 0)
	return windows, nil
}

// 辅助函数：直接调用 User32.dll
var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procEnumWindows          = user32.NewProc("EnumWindows")
)

// getWindowTextLength 返回窗口标题长度
func getWindowTextLength(hwnd win.HWND) int32 {
	ret, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	return int32(ret)
}

// getWindowText 将窗口标题写入缓冲区
func getWindowText(hwnd win.HWND, str *uint16, maxCount int32) int32 {
	ret, _, _ := procGetWindowTextW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(str)),
		uintptr(maxCount),
	)
	return int32(ret)
}

// enumWindows 枚举所有顶级窗口并调用回调
func enumWindows(lpEnumFunc uintptr, lParam uintptr) bool {
	ret, _, _ := procEnumWindows.Call(
		lpEnumFunc,
		lParam,
	)
	return ret != 0
}
