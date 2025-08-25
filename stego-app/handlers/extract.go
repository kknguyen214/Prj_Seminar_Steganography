package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

type ExtractRequest struct {
	Image      image.Image
	VideoData  []byte
	AudioData  []byte
	PDFData    []byte
	Passphrase string
	MediaType  string // "image", "video", "audio", "pdf"
}

type ExtractResponse struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message,omitempty"`
	MessageType string      `json:"message_type,omitempty"`
	Content     interface{} `json:"content,omitempty"`
	Timestamp   int64       `json:"timestamp,omitempty"`
	Size        int         `json:"size,omitempty"`
}

// ExtractHandler main API handler for extracting hidden messages
func ExtractHandler(c *gin.Context) {
	req, err := parseExtractRequest(c)
	if err != nil {
		respondExtractError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := processExtract(req)
	if err != nil {
		respondExtractError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// parseExtractRequest parses the extraction request
func parseExtractRequest(c *gin.Context) (*ExtractRequest, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("failed to parse multipart form")
	}

	req := &ExtractRequest{
		Passphrase: c.PostForm("passphrase"),
		MediaType:  c.PostForm("media_type"),
	}

	// Validate required fields
	if req.Passphrase == "" {
		return nil, errors.New("passphrase is required")
	}
	if req.MediaType == "" {
		return nil, errors.New("media_type is required (image/video/audio/pdf)")
	}

	// Parse media file containing hidden data
	if err := parseExtractMedia(form, req); err != nil {
		return nil, err
	}

	return req, nil
}

func parseExtractMedia(form *multipart.Form, req *ExtractRequest) error {
	var files []*multipart.FileHeader
	var fieldName string

	switch req.MediaType {
	case "image":
		files = form.File["image"]
		fieldName = "image"
	case "video":
		files = form.File["video"]
		fieldName = "video"
	case "audio":
		files = form.File["audio"]
		fieldName = "audio"
	case "pdf":
		files = form.File["pdf"]
		fieldName = "pdf"
	default:
		return errors.New("invalid media_type. Must be: image, video, audio, or pdf")
	}

	if len(files) == 0 {
		return errors.New("no " + fieldName + " file provided")
	}

	file := files[0]

	// Validate file format
	if !isValidExtractFormat(file.Filename, req.MediaType) {
		return errors.New("unsupported " + req.MediaType + " format")
	}

	src, err := file.Open()
	if err != nil {
		return errors.New("failed to open " + fieldName + " file")
	}
	defer src.Close()

	switch req.MediaType {
	case "image":
		img, _, err := image.Decode(src)
		if err != nil {
			return errors.New("invalid image format or corrupted file")
		}
		req.Image = img
	case "video":
		data, err := io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read video file")
		}
		req.VideoData = data
	case "audio":
		data, err := io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read audio file")
		}
		req.AudioData = data
	case "pdf":
		data, err := io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read pdf file")
		}
		req.PDFData = data
	}

	return nil
}

// processExtract handles the complete extraction process
func processExtract(req *ExtractRequest) (*ExtractResponse, error) {
	// Extract raw data from media
	var rawData []byte
	var err error

	switch req.MediaType {
	case "image":
		rawData, err = utils.ExtractDataFromImage(req.Image)
	case "video":
		rawData, err = utils.ExtractDataFromVideo(req.VideoData)
	case "audio":
		rawData, err = utils.ExtractDataFromAudio(req.AudioData)
	case "pdf":
		rawData, err = utils.ExtractDataFromPDF(req.PDFData)
	default:
		return nil, errors.New("invalid media type")
	}

	if err != nil {
		return nil, errors.New("failed to extract data from " + req.MediaType + ": " + err.Error())
	}

	// Check minimum data length (salt + some encrypted data)
	if len(rawData) < 32 { // 16 bytes salt + minimum encrypted data
		return nil, errors.New("no hidden data found or file is corrupted")
	}

	// Extract salt and encrypted data
	salt := rawData[:16]
	encrypted := rawData[16:]

	// Generate decryption key
	key := utils.GenerateKey(req.Passphrase, salt)

	// Decrypt the data
	decrypted, err := utils.DecryptData(encrypted, key)
	if err != nil {
		return nil, errors.New("invalid passphrase or corrupted encrypted data")
	}

	// Parse JSON message data
	var messageData map[string]interface{}
	if err := json.Unmarshal(decrypted, &messageData); err != nil {
		return nil, errors.New("corrupted message data - invalid JSON format")
	}

	// Extract message type
	messageType, ok := messageData["type"].(string)
	if !ok {
		return nil, errors.New("invalid message format - missing or invalid type")
	}

	// Extract content
	content, ok := messageData["content"].(string)
	if !ok {
		return nil, errors.New("invalid message format - missing or invalid content")
	}

	// Create response
	response := &ExtractResponse{
		Success:     true,
		MessageType: messageType,
	}

	// Extract optional fields
	if timestamp, ok := messageData["timestamp"].(float64); ok {
		response.Timestamp = int64(timestamp)
	}
	if size, ok := messageData["size"].(float64); ok {
		response.Size = int(size)
	}

	// Process content based on message type
	switch messageType {
	case "text":
		response.Content = content

	case "audio":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, errors.New("failed to decode audio content")
		}
		response.Content = map[string]interface{}{
			"data":     base64.StdEncoding.EncodeToString(decoded),
			"filename": "extracted_audio.wav",
			"size":     len(decoded),
		}

	case "image":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, errors.New("failed to decode image content")
		}
		response.Content = map[string]interface{}{
			"data":     base64.StdEncoding.EncodeToString(decoded),
			"filename": "extracted_image.png",
			"size":     len(decoded),
		}

	case "video":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, errors.New("failed to decode video content")
		}
		response.Content = map[string]interface{}{
			"data":     base64.StdEncoding.EncodeToString(decoded),
			"filename": "extracted_video.mp4",
			"size":     len(decoded),
		}
	case "pdf":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, errors.New("failed to decode pdf content")
		}
		response.Content = map[string]interface{}{
			"data":     base64.StdEncoding.EncodeToString(decoded),
			"filename": "extracted_pdf.pdf",
			"size":     len(decoded),
		}

	default:
		return nil, errors.New("unknown message type: " + messageType)
	}

	return response, nil
}

// isValidExtractFormat validates file format for extraction
func isValidExtractFormat(filename, mediaType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	switch mediaType {
	case "image":
		return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" || ext == ".tiff"
	case "video":
		return ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".mov" || ext == ".wmv" || ext == ".flv"
	case "audio":
		return ext == ".wav" || ext == ".mp3" || ext == ".flac" || ext == ".aac" || ext == ".ogg"
	case "pdf":
		return ext == ".pdf"
	}

	return false
}

// respondExtractError helper function for extraction error responses
func respondExtractError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ExtractResponse{
		Success: false,
		Message: message,
	})
}
