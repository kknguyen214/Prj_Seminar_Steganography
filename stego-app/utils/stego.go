package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
)

// Constants for steganography
const (
	// LSB bit positions for different channels
	LSBMask     = 0xFE // 11111110 - mask to clear LSB
	ExtractMask = 0x01 // 00000001 - mask to extract LSB

	// Magic number to identify embedded data
	MagicNumber = uint32(0xDEADBEEF)

	// Maximum data size (in bytes) that can be embedded
	MaxDataSize = 1024 * 1024 * 10 // 10MB

	// PBKDF2 iterations
	PBKDF2Iterations = 100000

	// Key size for AES-256
	KeySize = 32

	// Salt size
	SaltSize = 16
)

// // GenerateKey generates encryption key from passphrase and salt using PBKDF2
// func GenerateKey(passphrase string, salt []byte) []byte {
// 	if len(salt) != SaltSize {
// 		// Ensure salt is exactly 16 bytes
// 		paddedSalt := make([]byte, SaltSize)
// 		copy(paddedSalt, salt)
// 		salt = paddedSalt
// 	}
// 	return pbkdf2.Key([]byte(passphrase), salt, PBKDF2Iterations, KeySize, sha3.New256)
// }

// // EncryptData encrypts data using AES-256-GCM
// func EncryptData(data []byte, key []byte) ([]byte, error) {
// 	if len(key) != KeySize {
// 		return nil, fmt.Errorf("key must be %d bytes for AES-256", KeySize)
// 	}

// 	if len(data) == 0 {
// 		return nil, errors.New("data cannot be empty")
// 	}

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create cipher: %w", err)
// 	}

// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create GCM: %w", err)
// 	}

// 	nonce := make([]byte, gcm.NonceSize())
// 	if _, err := rand.Read(nonce); err != nil {
// 		return nil, fmt.Errorf("failed to generate nonce: %w", err)
// 	}

// 	ciphertext := gcm.Seal(nonce, nonce, data, nil)
// 	return ciphertext, nil
// }

// DecryptData decrypts data using AES-256-GCM
func DecryptData(encryptedData []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes for AES-256", KeySize)
	}

	if len(encryptedData) == 0 {
		return nil, errors.New("encrypted data cannot be empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// EmbedDataInImage embeds data into an image using LSB steganography
func EmbedDataInImage(img image.Image, data []byte) ([]byte, error) {
	if img == nil {
		return nil, errors.New("image cannot be nil")
	}

	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid image dimensions")
	}

	// Calculate available capacity (3 channels * width * height bits / 8)
	// Reserve some space for padding and error correction
	capacity := (width * height * 3) / 8

	// Prepare data with length prefix and magic number
	dataWithHeader := prepareDataWithHeader(data)

	if len(dataWithHeader) > capacity {
		return nil, fmt.Errorf("image too small to embed data: need %d bytes, have %d bytes capacity",
			len(dataWithHeader), capacity)
	}

	// Create new RGBA image
	newImg := image.NewRGBA(bounds)

	// Convert data to bits
	dataBits := bytesToBits(dataWithHeader)
	bitIndex := 0

	// Embed data using LSB
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldPixel := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			newPixel := oldPixel

			// Embed in R channel
			if bitIndex < len(dataBits) {
				newPixel.R = (oldPixel.R & LSBMask) | dataBits[bitIndex]
				bitIndex++
			}

			// Embed in G channel
			if bitIndex < len(dataBits) {
				newPixel.G = (oldPixel.G & LSBMask) | dataBits[bitIndex]
				bitIndex++
			}

			// Embed in B channel
			if bitIndex < len(dataBits) {
				newPixel.B = (oldPixel.B & LSBMask) | dataBits[bitIndex]
				bitIndex++
			}

			newImg.SetRGBA(x, y, newPixel)

			// Break if all data is embedded
			if bitIndex >= len(dataBits) {
				// Fill remaining pixels with original data
				for yy := y; yy < bounds.Max.Y; yy++ {
					startX := bounds.Min.X
					if yy == y {
						startX = x + 1
					}
					for xx := startX; xx < bounds.Max.X; xx++ {
						originalPixel := color.RGBAModel.Convert(img.At(xx, yy)).(color.RGBA)
						newImg.SetRGBA(xx, yy, originalPixel)
					}
				}
				goto embedComplete
			}
		}
	}

embedComplete:
	// Encode to PNG
	var buf bytes.Buffer
	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	if err := encoder.Encode(&buf, newImg); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// ExtractDataFromImage extracts data from an image using LSB steganography
func ExtractDataFromImage(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, errors.New("image cannot be nil")
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid image dimensions")
	}

	var extractedBits []uint8

	// Extract LSBs from each channel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)

			// Extract from R, G, B channels
			extractedBits = append(extractedBits, pixel.R&ExtractMask)
			extractedBits = append(extractedBits, pixel.G&ExtractMask)
			extractedBits = append(extractedBits, pixel.B&ExtractMask)

			// Check if we have enough bits for header
			if len(extractedBits) >= 64 { // 8 bytes * 8 bits = 64 bits for header
				// Try to extract header at 64-bit boundaries
				if len(extractedBits)%64 == 0 {
					headerBytes := bitsToBytes(extractedBits[:64])
					if len(headerBytes) >= 8 {
						magic := binary.LittleEndian.Uint32(headerBytes[:4])
						if magic == MagicNumber {
							dataLength := binary.LittleEndian.Uint32(headerBytes[4:8])
							if dataLength > 0 && dataLength <= MaxDataSize {
								totalBitsNeeded := 64 + int(dataLength)*8

								// Continue extracting if we don't have enough bits yet
								if len(extractedBits) >= totalBitsNeeded {
									dataBytes := bitsToBytes(extractedBits[64:totalBitsNeeded])
									return dataBytes, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, errors.New("no valid embedded data found in image")
}

// EmbedDataInVideo embeds data into video file
func EmbedDataInVideo(videoData []byte, data []byte) ([]byte, error) {
	if len(videoData) == 0 {
		return nil, errors.New("video data cannot be empty")
	}

	if len(data) == 0 {
		return nil, errors.New("data to embed cannot be empty")
	}

	if len(data) > MaxDataSize {
		return nil, fmt.Errorf("data too large: %d bytes, max allowed: %d bytes", len(data), MaxDataSize)
	}

	// For a complete implementation, you would parse video container format
	// This is a simplified approach: append encrypted data at the end
	dataWithHeader := prepareDataWithHeader(data)

	// Check if video file has enough space (heuristic check)
	if len(dataWithHeader) > len(videoData)/100 { // Data shouldn't be more than 1% of video size
		return nil, errors.New("data too large relative to video file size")
	}

	// Create new video data by appending our data
	result := make([]byte, len(videoData)+len(dataWithHeader))
	copy(result, videoData)
	copy(result[len(videoData):], dataWithHeader)

	return result, nil
}

// ExtractDataFromVideo extracts data from video file
func ExtractDataFromVideo(videoData []byte) ([]byte, error) {
	if len(videoData) < 8 {
		return nil, errors.New("video file too small")
	}

	// Look for magic number in the last part of file
	// Start search from end, looking backwards
	searchStart := len(videoData) - MaxDataSize - 8
	if searchStart < 0 {
		searchStart = 0
	}

	// Search for magic number
	for i := len(videoData) - 8; i >= searchStart; i-- {
		if i+8 <= len(videoData) {
			magic := binary.LittleEndian.Uint32(videoData[i : i+4])
			if magic == MagicNumber {
				dataLength := binary.LittleEndian.Uint32(videoData[i+4 : i+8])
				if dataLength > 0 && dataLength <= MaxDataSize {
					totalLength := 8 + int(dataLength)
					if i+totalLength <= len(videoData) {
						return videoData[i+8 : i+totalLength], nil
					}
				}
			}
		}
	}

	return nil, errors.New("no embedded data found in video")
}

// EmbedDataInAudio embeds data into audio file using LSB in audio samples
func EmbedDataInAudio(audioData []byte, data []byte) ([]byte, error) {
	if len(audioData) == 0 {
		return nil, errors.New("audio data cannot be empty")
	}

	if len(data) == 0 {
		return nil, errors.New("data to embed cannot be empty")
	}

	// Basic WAV file validation
	if len(audioData) < 44 { // Minimum WAV header size
		return nil, errors.New("audio file too small or invalid format")
	}

	// Check for WAV header
	if string(audioData[:4]) != "RIFF" || string(audioData[8:12]) != "WAVE" {
		return nil, errors.New("unsupported audio format - only WAV files supported for LSB embedding")
	}

	// Find data chunk
	headerSize := findWavDataChunk(audioData)
	if headerSize == -1 {
		return nil, errors.New("invalid WAV file - no data chunk found")
	}

	audioSamples := audioData[headerSize:]
	dataWithHeader := prepareDataWithHeader(data)

	// Calculate capacity (1 bit per audio sample byte)
	capacity := len(audioSamples) / 8 // 1 bit per byte = 1/8 capacity

	if len(dataWithHeader) > capacity {
		return nil, fmt.Errorf("audio file too small to embed data: need %d bytes, have %d bytes capacity",
			len(dataWithHeader), capacity)
	}

	// Create new audio data
	result := make([]byte, len(audioData))
	copy(result, audioData)

	// Convert data to bits
	dataBits := bytesToBits(dataWithHeader)

	// Embed data in audio samples using LSB
	for i, bit := range dataBits {
		if headerSize+i < len(result) {
			result[headerSize+i] = (result[headerSize+i] & LSBMask) | bit
		}
	}

	return result, nil
}

// ExtractDataFromAudio extracts data from audio file
func ExtractDataFromAudio(audioData []byte) ([]byte, error) {
	if len(audioData) < 44 {
		return nil, errors.New("audio file too small or invalid")
	}

	// Find WAV data chunk
	headerSize := findWavDataChunk(audioData)
	if headerSize == -1 {
		// For non-WAV files, try with default header size
		headerSize = 44
	}

	if len(audioData) <= headerSize {
		return nil, errors.New("audio file has no sample data")
	}

	audioSamples := audioData[headerSize:]

	if len(audioSamples) < 64 { // Need at least 8 bytes for header
		return nil, errors.New("audio file too small to contain embedded data")
	}

	var extractedBits []uint8

	// Extract LSBs from audio samples
	for i, sample := range audioSamples {
		extractedBits = append(extractedBits, sample&ExtractMask)

		// Check for header every 64 bits
		if len(extractedBits) >= 64 && len(extractedBits)%64 == 0 {
			headerBytes := bitsToBytes(extractedBits[:64])
			if len(headerBytes) >= 8 {
				magic := binary.LittleEndian.Uint32(headerBytes[:4])
				if magic == MagicNumber {
					dataLength := binary.LittleEndian.Uint32(headerBytes[4:8])
					if dataLength > 0 && dataLength <= MaxDataSize {
						totalBitsNeeded := 64 + int(dataLength)*8

						// Check if we have enough samples left
						if totalBitsNeeded <= len(audioSamples) {
							if len(extractedBits) >= totalBitsNeeded {
								dataBytes := bitsToBytes(extractedBits[64:totalBitsNeeded])
								return dataBytes, nil
							}
						} else {
							// Not enough samples, break early
							break
						}
					}
				}
			}
		}

		// Safety check to prevent excessive memory usage
		if i > MaxDataSize*8 {
			break
		}
	}

	return nil, errors.New("no embedded data found in audio")
}

// Helper functions

// findWavDataChunk finds the data chunk in a WAV file and returns its offset
func findWavDataChunk(wavData []byte) int {
	if len(wavData) < 12 {
		return -1
	}

	// Skip RIFF header (12 bytes)
	offset := 12

	for offset+8 <= len(wavData) {
		chunkID := string(wavData[offset : offset+4])
		chunkSize := binary.LittleEndian.Uint32(wavData[offset+4 : offset+8])

		if chunkID == "data" {
			return offset + 8
		}

		// Move to next chunk
		offset += 8 + int(chunkSize)

		// Add padding byte if chunk size is odd
		if chunkSize%2 == 1 {
			offset++
		}
	}

	return -1
}

// prepareDataWithHeader adds magic number and length header to data
func prepareDataWithHeader(data []byte) []byte {
	if len(data) > MaxDataSize {
		// Truncate data if too large
		data = data[:MaxDataSize]
	}

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[:4], MagicNumber)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(data)))

	result := make([]byte, len(header)+len(data))
	copy(result, header)
	copy(result[len(header):], data)

	return result
}

// bytesToBits converts bytes to individual bits (LSB first)
func bytesToBits(data []byte) []uint8 {
	var bits []uint8
	for _, b := range data {
		for i := 0; i < 8; i++ {
			bits = append(bits, (b>>i)&1)
		}
	}
	return bits
}

// bitsToBytes converts individual bits to bytes (LSB first)
func bitsToBytes(bits []uint8) []byte {
	if len(bits) == 0 {
		return []byte{}
	}

	// Pad with zeros if necessary
	for len(bits)%8 != 0 {
		bits = append(bits, 0)
	}

	var bytes []byte
	for i := 0; i < len(bits); i += 8 {
		var b uint8
		for j := 0; j < 8 && i+j < len(bits); j++ {
			if bits[i+j] == 1 {
				b |= 1 << j
			}
		}
		bytes = append(bytes, b)
	}
	return bytes
}

// calculateImageCapacity calculates how many bytes can be embedded in an image
func CalculateImageCapacity(width, height int) int {
	if width <= 0 || height <= 0 {
		return 0
	}
	// 3 channels * width * height bits / 8, minus header overhead
	capacity := (width * height * 3) / 8
	return int(math.Max(0, float64(capacity-8))) // Reserve 8 bytes for header
}

// calculateAudioCapacity calculates how many bytes can be embedded in audio
func CalculateAudioCapacity(audioDataSize int) int {
	if audioDataSize <= 44 {
		return 0
	}
	// 1 bit per sample byte, minus header size and overhead
	capacity := (audioDataSize - 44) / 8
	return int(math.Max(0, float64(capacity-8))) // Reserve 8 bytes for header
}

// validateImageForSteganography checks if image is suitable for steganography
func ValidateImageForSteganography(img image.Image, dataSize int) error {
	if img == nil {
		return errors.New("image is nil")
	}

	bounds := img.Bounds()
	capacity := CalculateImageCapacity(bounds.Dx(), bounds.Dy())

	if dataSize > capacity {
		return fmt.Errorf("image too small: need %d bytes capacity, have %d bytes", dataSize, capacity)
	}

	return nil
}

// validateAudioForSteganography checks if audio file is suitable for steganography
func ValidateAudioForSteganography(audioDataSize, dataSize int) error {
	capacity := CalculateAudioCapacity(audioDataSize)

	if dataSize > capacity {
		return fmt.Errorf("audio file too small: need %d bytes capacity, have %d bytes", dataSize, capacity)
	}

	return nil
}
