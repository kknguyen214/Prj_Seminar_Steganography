package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/auyer/steganography"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/pbkdf2"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ExtractedData struct {
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
}

// Generate key from passphrase using PBKDF2
func generateKey(passphrase string, salt []byte) []byte {
	return pbkdf2.Key([]byte(passphrase), salt, 10000, 32, sha256.New)
}

// Encrypt data using AES-GCM
func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt data using AES-GCM
func decryptData(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Validate and get supported image format
func getImageFormat(filename string) string {
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

// Convert image to NRGBA format
func imageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)
	draw.Draw(nrgbaImg, bounds, img, bounds.Min, draw.Src)
	return nrgbaImg
}

// Embed data into image using steganography library
func embedDataInImage(img image.Image, data []byte) (*bytes.Buffer, error) {
	// Convert to NRGBA
	nrgbaImg := imageToNRGBA(img)

	// Check if image has enough capacity
	maxSize := steganography.GetMessageSizeFromImage(nrgbaImg)
	if uint32(len(data)) > maxSize {
		return nil, fmt.Errorf("image too small to embed data - max capacity: %d bytes, data size: %d bytes", maxSize, len(data))
	}

	// Create buffer for output
	var buf bytes.Buffer

	// Use the steganography library to embed data
	steganography.EncodeNRGBA(&buf, nrgbaImg, data)

	return &buf, nil
}

// Extract data from image using steganography library
func extractDataFromImage(img image.Image) ([]byte, error) {
	// Convert to NRGBA
	nrgbaImg := imageToNRGBA(img)

	// Get the message size from the image
	sizeOfMessage := steganography.GetMessageSizeFromImage(nrgbaImg)

	if sizeOfMessage == 0 {
		return nil, fmt.Errorf("no hidden data found")
	}

	// Extract the hidden message
	data := steganography.Decode(sizeOfMessage, nrgbaImg)

	return data, nil
}

func embedHandler(c *gin.Context) {
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Failed to parse form data",
		})
		return
	}

	// Get image file
	files := form.File["image"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "No image file provided",
		})
		return
	}

	file := files[0]

	// Validate file format before processing
	supportedFormat := getImageFormat(file.Filename)
	if supportedFormat == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Unsupported image format. File: %s. Supported: .jpg, .jpeg, .png, .gif, .bmp, .tiff, .webp", file.Filename),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to open image file",
		})
		return
	}
	defer src.Close()

	// Get file info for debugging
	contentType := file.Header.Get("Content-Type")
	filename := file.Filename
	fileSize := file.Size

	fmt.Printf("Uploaded file: %s, Content-Type: %s, Size: %d bytes\n",
		filename, contentType, fileSize)

	// Reset file position
	src.Seek(0, 0)

	// Try to decode image with better error handling
	img, format, err := image.Decode(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid image format - %s (file: %s, type: %s)",
				err.Error(), filename, contentType),
		})
		return
	}

	fmt.Printf("Decoded image format: %s, bounds: %v\n", format, img.Bounds())

	// Get form parameters
	passphrase := c.PostForm("passphrase")
	messageType := c.PostForm("message_type")
	text := c.PostForm("text")

	if passphrase == "" || messageType == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Passphrase and message_type are required",
		})
		return
	}

	// Prepare message data
	messageData := map[string]interface{}{
		"type":    messageType,
		"content": text,
	}

	// Handle different message types
	switch messageType {
	case "text":
		if text == "" {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Text content is required for text message type",
			})
			return
		}
	case "audio":
		audioFiles := form.File["audio"]
		if len(audioFiles) == 0 {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Audio file is required for audio message type",
			})
			return
		}
		audioFile := audioFiles[0]
		audioSrc, err := audioFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to open audio file",
			})
			return
		}
		defer audioSrc.Close()

		audioData, err := io.ReadAll(audioSrc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to read audio file",
			})
			return
		}
		messageData["content"] = base64.StdEncoding.EncodeToString(audioData)

	case "image":
		msgImageFiles := form.File["message_image"]
		if len(msgImageFiles) == 0 {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Message image file is required for image message type",
			})
			return
		}
		msgImageFile := msgImageFiles[0]
		msgImageSrc, err := msgImageFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to open message image file",
			})
			return
		}
		defer msgImageSrc.Close()

		msgImageData, err := io.ReadAll(msgImageSrc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to read message image file",
			})
			return
		}
		messageData["content"] = base64.StdEncoding.EncodeToString(msgImageData)
	}

	// Serialize message data
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to serialize message data",
		})
		return
	}

	// Generate salt and key
	salt := make([]byte, 16)
	rand.Read(salt)
	key := generateKey(passphrase, salt)

	// Encrypt data
	encryptedData, err := encryptData(jsonData, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to encrypt data",
		})
		return
	}

	// Prepend salt to encrypted data
	fullData := append(salt, encryptedData...)

	// Embed data in image
	resultBuffer, err := embedDataInImage(img, fullData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Set response headers and return the image
	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", "attachment; filename=embedded_image.png")
	c.Data(http.StatusOK, "image/png", resultBuffer.Bytes())
}

func extractHandler(c *gin.Context) {
	// Get image file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "No image file provided",
		})
		return
	}

	// Validate file format
	supportedFormat := getImageFormat(file.Filename)
	if supportedFormat == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Unsupported image format. File: %s", file.Filename),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to open image file",
		})
		return
	}
	defer src.Close()

	// Get file info for debugging
	fmt.Printf("Extracting from file: %s, Size: %d bytes\n",
		file.Filename, file.Size)

	// Try to decode image with better error handling
	img, format, err := image.Decode(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid image format - %s (file: %s)",
				err.Error(), file.Filename),
		})
		return
	}

	fmt.Printf("Decoded image format: %s, bounds: %v\n", format, img.Bounds())

	// Get passphrase
	passphrase := c.PostForm("passphrase")
	if passphrase == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Passphrase is required",
		})
		return
	}

	// Extract data from image
	extractedData, err := extractDataFromImage(img)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Extract salt and encrypted data
	if len(extractedData) < 16 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid embedded data format",
		})
		return
	}

	salt := extractedData[:16]
	encryptedData := extractedData[16:]

	// Generate key and decrypt
	key := generateKey(passphrase, salt)
	decryptedData, err := decryptData(encryptedData, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Failed to decrypt data - wrong passphrase?",
		})
		return
	}

	// Parse decrypted JSON
	var messageData map[string]interface{}
	err = json.Unmarshal(decryptedData, &messageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to parse decrypted data",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: ExtractedData{
			MessageType: messageData["type"].(string),
			Content:     messageData["content"].(string),
		},
	})
}

func main() {
	r := gin.Default()

	// Set max memory for multipart forms (32MB)
	r.MaxMultipartMemory = 32 << 20

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	r.POST("/embed", embedHandler)
	r.POST("/extract", extractHandler)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fmt.Println("Server starting on :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /embed   - Embed message into image")
	fmt.Println("  POST /extract - Extract message from image")
	fmt.Println("  GET  /health  - Health check")
	r.Run(":8080")
}
