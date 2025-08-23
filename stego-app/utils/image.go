package utils

import (
	"image"
	"image/draw"
	"path/filepath"
	"strings"
)

// GetImageFormat kiểm tra format ảnh
func GetImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".bmp":
		return "bmp"
	case ".tiff", ".tif":
		return "tiff"
	case ".webp":
		return "webp"
	default:
		return ""
	}
}

// ImageToNRGBA convert bất kỳ ảnh nào sang NRGBA
func ImageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)
	draw.Draw(nrgbaImg, bounds, img, bounds.Min, draw.Src)
	return nrgbaImg
}
