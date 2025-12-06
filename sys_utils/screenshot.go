package sys_utils

import (
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"cron-shot/utils"

	"github.com/lxn/win"
)

// CaptureWindowImage 使用 PrintWindow 渲染窗口；失败时记录日志，不进行回退
func CaptureWindowImage(hwnd win.HWND) (*image.RGBA, error) {
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	hdcScreen := win.GetDC(0)
	defer win.ReleaseDC(0, hdcScreen)
	hdcMem := win.CreateCompatibleDC(hdcScreen)
	defer win.DeleteDC(hdcMem)
	hbm := win.CreateCompatibleBitmap(hdcScreen, int32(width), int32(height))
	defer win.DeleteObject(win.HGDIOBJ(hbm))
	win.SelectObject(hdcMem, win.HGDIOBJ(hbm))
	const PW_RENDERFULLCONTENT = 0x00000002
	user32 := syscall.NewLazyDLL("user32.dll")
	printWindow := user32.NewProc("PrintWindow")
	r, _, _ := printWindow.Call(uintptr(hwnd), uintptr(hdcMem), uintptr(PW_RENDERFULLCONTENT))
	if r == 0 {
		return nil, errors.New("PrintWindow failed")
	}
	var bmi win.BITMAPINFO
	bmi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.BiWidth = int32(width)
	bmi.BmiHeader.BiHeight = -int32(height)
	bmi.BmiHeader.BiPlanes = 1
	bmi.BmiHeader.BiBitCount = 32
	bmi.BmiHeader.BiCompression = win.BI_RGB
	stride := width * 4
	buf := make([]byte, stride*height)
	win.GetDIBits(hdcMem, hbm, 0, uint32(height), &buf[0], &bmi, win.DIB_RGB_COLORS)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	pi := 0
	for y := 0; y < height; y++ {
		row := y * stride
		for x := 0; x < width; x++ {
			i := row + x*4
			b := buf[i+0]
			g := buf[i+1]
			r := buf[i+2]
			a := buf[i+3]
			if a == 0 {
				a = 255
			}
			img.Pix[pi+0] = r
			img.Pix[pi+1] = g
			img.Pix[pi+2] = b
			img.Pix[pi+3] = a
			pi += 4
		}
	}
	return img, nil
}

// SaveCronShot 保存截图到 根目录\\进程名\\(固定文件夹)\\规则文件夹 下
func SaveCronShot(img *image.RGBA, root string, processName, fixedFolder, folderName string, t time.Time) (string, error) {
	proc := utils.SanitizeProcessName(processName)
	sub := utils.SanitizeFolderName(folderName)
	if fixedFolder != "" {
		fix := utils.SanitizeFolderName(fixedFolder)
		dir := filepath.Join(root, proc, fix, sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
		name := t.Format("20060102_150405.000") + ".png"
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			return "", err
		}
		return path, nil
	}
	dir := filepath.Join(root, proc, sub)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := t.Format("20060102_150405.000") + ".png"
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return "", err
	}
	return path, nil
}
