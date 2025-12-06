package utils

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AHash16x16 生成16x16平均哈希（256位）
func AHash16x16(img *image.RGBA) []byte {
	w, h := 16, 16
	// 计算缩放步长
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return make([]byte, 32)
	}
	stepX := float64(srcW) / float64(w)
	stepY := float64(srcH) / float64(h)
	gray := make([]float64, w*h)
	// 采样像素并转灰度
	idx := 0
	var sum float64
	for y := 0; y < h; y++ {
		sy := int(float64(y)*stepY + stepY/2)
		if sy >= srcH {
			sy = srcH - 1
		}
		for x := 0; x < w; x++ {
			sx := int(float64(x)*stepX + stepX/2)
			if sx >= srcW {
				sx = srcW - 1
			}
			o := img.RGBAAt(img.Bounds().Min.X+sx, img.Bounds().Min.Y+sy)
			g := 0.2126*float64(o.R) + 0.7152*float64(o.G) + 0.0722*float64(o.B)
			gray[idx] = g
			sum += g
			idx++
		}
	}
	avg := sum / float64(w*h)
	bits := make([]byte, 32)
	for i := 0; i < w*h; i++ {
		if gray[i] >= avg {
			bits[i>>3] |= (1 << uint(7-(i&7)))
		}
	}
	return bits
}

// Hamming256 计算两个256位哈希的汉明距离
func Hamming256(a, b []byte) int {
	if len(a) != 32 || len(b) != 32 {
		return 256
	}
	dist := 0
	for i := 0; i < 32; i++ {
		x := a[i] ^ b[i]
		// 位计数
		x = (x & 0x55) + ((x >> 1) & 0x55)
		x = (x & 0x33) + ((x >> 2) & 0x33)
		x = (x & 0x0F) + ((x >> 4) & 0x0F)
		dist += int(x)
	}
	return dist
}

// AHashFromImage 生成任意 image.Image 的16x16平均哈希
func AHashFromImage(img image.Image) []byte {
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, img, b.Min, draw.Src)
	return AHash16x16(rgba)
}

// LatestPNGImage 返回目录下最新PNG文件的解码图像；不存在则返回nil
func LatestPNGImage(dir string) (image.Image, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var latestPath string
	var latestMod time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasSuffix(name, ".png") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestPath = filepath.Join(dir, e.Name())
		}
	}
	if latestPath == "" {
		return nil, nil
	}
	f, err := os.Open(latestPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// ImagesEqualExact 比较两张图是否像素完全一致（尺寸与像素均相同）
func ImagesEqualExact(a *image.RGBA, b image.Image) bool {
	bounds := b.Bounds()
	if a.Bounds().Dx() != bounds.Dx() || a.Bounds().Dy() != bounds.Dy() {
		return false
	}
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, b, bounds.Min, draw.Src)
	ap := a.Pix
	bp := rgba.Pix
	if len(ap) != len(bp) {
		return false
	}
	for i := 0; i < len(ap); i++ {
		if ap[i] != bp[i] {
			return false
		}
	}
	return true
}
