package sys_utils

import (
	"cron-shot/constants"
	"syscall"
	"unsafe"
)

var (
	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
)

type browseInfoW struct {
	HwndOwner      uintptr
	PidlRoot       uintptr
	PszDisplayName *uint16
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}

// PickFolder 调用 Windows 原生 SHBrowseForFolder 弹出文件夹选择对话框
func PickFolder() (string, error) {
	var bi browseInfoW
	title := utf16FromString(constants.TextChoose)
	bi.LpszTitle = &title[0]
	bi.UlFlags = 0x0040 | 0x0001 // BIF_NEWDIALOGSTYLE | BIF_RETURNONLYFSDIRS
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil
	}
	defer procCoTaskMemFree.Call(pidl)

	buf := make([]uint16, 260)
	ok, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	if ok == 0 {
		return "", nil
	}
	return syscall.UTF16ToString(buf), nil
}

// utf16FromString 将 Go 字符串转换为 UTF-16 切片
func utf16FromString(s string) []uint16 {
	u := syscall.StringToUTF16(s)
	return u
}
