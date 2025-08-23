package utils

import (
	"bytes"
	"fmt"
	"image"

	"github.com/auyer/steganography"
)

// EmbedDataInImage
func EmbedDataInImage(img image.Image, data []byte) (*bytes.Buffer, error) {
	nrgba := ImageToNRGBA(img)
	maxSize := steganography.GetMessageSizeFromImage(nrgba)
	if uint32(len(data)) > maxSize {
		return nil, fmt.Errorf("image too small to embed data (%d bytes max, %d bytes given)", maxSize, len(data))
	}
	buf := new(bytes.Buffer)
	steganography.EncodeNRGBA(buf, nrgba, data)
	return buf, nil
}

// ExtractDataFromImage
func ExtractDataFromImage(img image.Image) ([]byte, error) {
	nrgba := ImageToNRGBA(img)
	size := steganography.GetMessageSizeFromImage(nrgba)
	if size == 0 {
		return nil, fmt.Errorf("no hidden data found")
	}
	data := steganography.Decode(size, nrgba)
	return data, nil
}
