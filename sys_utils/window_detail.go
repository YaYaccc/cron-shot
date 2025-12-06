package sys_utils

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type WindowInfo struct {
	Title string
	HWND  win.HWND
}

// GetProcessWindowsDetailed 返回指定进程的可见窗口详细信息（标题与句柄）
func GetProcessWindowsDetailed(processName string) ([]WindowInfo, error) {
	var result []WindowInfo
	target := processName
	if !strings.HasSuffix(strings.ToLower(target), ".exe") {
		target += ".exe"
	}

	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		var pid uint32
		win.GetWindowThreadProcessId(hwnd, &pid)

		const PROCESS_QUERY_INFORMATION = 0x0400
		const PROCESS_VM_READ = 0x0010
		hProcess, err := syscall.OpenProcess(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, false, pid)
		if err == nil && hProcess != 0 {
			defer syscall.CloseHandle(hProcess)

			var nameBuf [win.MAX_PATH]uint16
			n := uint32(len(nameBuf))
			ret, _, _ := syscall.NewLazyDLL("psapi.dll").NewProc("GetModuleBaseNameW").Call(
				uintptr(hProcess),
				0,
				uintptr(unsafe.Pointer(&nameBuf[0])),
				uintptr(n),
			)
			if ret != 0 {
				pName := syscall.UTF16ToString(nameBuf[:])
				if strings.EqualFold(pName, target) {
					if win.IsWindowVisible(hwnd) {
						tl := getWindowTextLength(hwnd)
						if tl > 0 {
							buf := make([]uint16, tl+1)
							getWindowText(hwnd, &buf[0], int32(tl+1))
							title := syscall.UTF16ToString(buf)
							if title != "" {
								result = append(result, WindowInfo{Title: title, HWND: hwnd})
							}
						}
					}
				}
			}
		}
		return 1
	})

	enumWindows(cb, 0)
	return result, nil
}
