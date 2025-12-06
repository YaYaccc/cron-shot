package sys_utils

import (
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	shell32                  = syscall.NewLazyDLL("shell32.dll")
	ole32                    = syscall.NewLazyDLL("ole32.dll")
	procSHGetKnownFolderPath = shell32.NewProc("SHGetKnownFolderPath")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	FOLDERID_Pictures = GUID{
		Data1: 0x33E28130,
		Data2: 0x4E1E,
		Data3: 0x4676,
		Data4: [8]byte{0x83, 0x5A, 0x98, 0x39, 0x5C, 0x3B, 0xC3, 0xBB},
	}
)

// SHGetKnownFolderPath 通过 GUID 获取系统已知文件夹路径
func SHGetKnownFolderPath(folderID *GUID) (string, error) {
	var pPath uintptr
	ret, _, _ := procSHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(folderID)),
		0,
		0,
		uintptr(unsafe.Pointer(&pPath)),
	)
	if ret != 0 {
		return "", syscall.Errno(ret)
	}
	defer procCoTaskMemFree.Call(pPath)
	return utf16PtrToString((*uint16)(unsafe.Pointer(pPath))), nil
}

// GetWindowsPicturesFolder 返回 Windows "图片" 文件夹路径
func GetWindowsPicturesFolder() (string, error) {
	return SHGetKnownFolderPath(&FOLDERID_Pictures)
}

// GetPicturesFolderWithFallback 优先返回系统图片目录，失败时回退到用户家目录下的 Pictures
func GetPicturesFolderWithFallback() string {
	p, err := GetWindowsPicturesFolder()
	if err == nil && p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return home + string(os.PathSeparator) + "Pictures"
}

// utf16PtrToString 从 UTF-16 指针构造 Go 字符串
func utf16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	length := 0
	for p := ptr; *p != 0; p = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + 2)) {
		length++
	}
	utf16Slice := make([]uint16, length)
	for i := 0; i < length; i++ {
		utf16Slice[i] = *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(i*2)))
	}
	return string(utf16.Decode(utf16Slice))
}
