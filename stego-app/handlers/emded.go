package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg" // Nhận jpeg
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

// EmbedRequest represents the request for embedding secret message into media
type EmbedRequest struct {
	// Carrier media files (where to embed into)
	Image     image.Image
	VideoData []byte
	AudioData []byte

	// Metadata
	Passphrase  string
	MediaType   string // "image", "video", "audio" - carrier media type
	MessageType string // "text", "audio", "image", "video" - secret message type

	// Secret message content
	Text         string
	MessageAudio []byte
	MessageImage []byte
	MessageVideo []byte

	// Original filename for proper response
	OriginalFilename string
}

// EmbedResponse represents the response structure
type EmbedResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// EmbedHandler handles HTTP request for embedding secret message into media
func EmbedHandler(c *gin.Context) {
	// Parse request
	req, err := parseEmbedRequest(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Process embedding
	result, contentType, filename, err := processEmbed(req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return result with proper headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, result)
}

// parseEmbedRequest parses all request data
func parseEmbedRequest(c *gin.Context) (*EmbedRequest, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("failed to parse multipart form")
	}

	req := &EmbedRequest{
		Passphrase:  c.PostForm("passphrase"),
		MediaType:   c.PostForm("media_type"),   // "image", "video", "audio" - carrier
		MessageType: c.PostForm("message_type"), // "text", "audio", "image", "video" - secret
		Text:        c.PostForm("text"),
	}

	// Validate required fields
	if req.Passphrase == "" {
		return nil, errors.New("passphrase is required")
	}
	if req.MediaType == "" {
		return nil, errors.New("media_type is required (image/video/audio)")
	}
	if req.MessageType == "" {
		return nil, errors.New("message_type is required (text/audio/image/video)")
	}

	// Parse carrier media file (where to embed into)
	if err := parseCarrierMedia(form, req); err != nil {
		return nil, err
	}

	// Parse secret message content
	if err := parseMessageContent(form, req); err != nil {
		return nil, err
	}

	return req, nil
}

// parseCarrierMedia parses the carrier media file (image/video/audio)
func parseCarrierMedia(form *multipart.Form, req *EmbedRequest) error {
	var files []*multipart.FileHeader
	var fieldName string

	switch req.MediaType {
	case "image":
		files = form.File["carrier_image"]
		fieldName = "carrier_image"
	case "video":
		files = form.File["carrier_video"]
		fieldName = "carrier_video"
	case "audio":
		files = form.File["carrier_audio"]
		fieldName = "carrier_audio"
	default:
		return errors.New("invalid media_type. Must be: image, video, or audio")
	}

	if len(files) == 0 {
		return errors.New("no " + fieldName + " file provided")
	}

	file := files[0]
	req.OriginalFilename = file.Filename

	// Validate file format
	if !isValidCarrierFormat(file.Filename, req.MediaType) {
		return errors.New("unsupported " + req.MediaType + " format")
	}

	src, err := file.Open()
	if err != nil {
		return errors.New("failed to open " + fieldName + " file")
	}
	defer src.Close()

	switch req.MediaType {
	case "image":
		// For images, decode to image.Image
		img, _, err := image.Decode(src)
		if err != nil {
			return errors.New("invalid image format or corrupted file")
		}
		req.Image = img
	case "video":
		// For video, read as bytes
		data, err := io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read video file")
		}
		req.VideoData = data
	case "audio":
		// For audio, read as bytes
		data, err := io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read audio file")
		}
		req.AudioData = data
	}

	return nil
}

// parseMessageContent parses the secret message content
func parseMessageContent(form *multipart.Form, req *EmbedRequest) error {
	switch req.MessageType {
	case "text":
		if req.Text == "" {
			return errors.New("text content is required for text message type")
		}
		return nil

	case "audio":
		audioFiles := form.File["message_audio"]
		if len(audioFiles) == 0 {
			return errors.New("message audio file is required for audio message type")
		}

		if !isValidMessageFormat(audioFiles[0].Filename, "audio") {
			return errors.New("unsupported message audio format")
		}

		src, err := audioFiles[0].Open()
		if err != nil {
			return errors.New("failed to open message audio file")
		}
		defer src.Close()

		req.MessageAudio, err = io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read message audio file")
		}
		return nil

	case "image":
		imgFiles := form.File["message_image"]
		if len(imgFiles) == 0 {
			return errors.New("message image file is required for image message type")
		}

		if !isValidMessageFormat(imgFiles[0].Filename, "image") {
			return errors.New("unsupported message image format")
		}

		src, err := imgFiles[0].Open()
		if err != nil {
			return errors.New("failed to open message image file")
		}
		defer src.Close()

		req.MessageImage, err = io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read message image file")
		}
		return nil

	case "video":
		videoFiles := form.File["message_video"]
		if len(videoFiles) == 0 {
			return errors.New("message video file is required for video message type")
		}

		if !isValidMessageFormat(videoFiles[0].Filename, "video") {
			return errors.New("unsupported message video format")
		}

		src, err := videoFiles[0].Open()
		if err != nil {
			return errors.New("failed to open message video file")
		}
		defer src.Close()

		req.MessageVideo, err = io.ReadAll(src)
		if err != nil {
			return errors.New("failed to read message video file")
		}
		return nil

	default:
		return errors.New("invalid message_type. Must be: text, audio, image, or video")
	}
}

// isValidCarrierFormat validates carrier media file format
func isValidCarrierFormat(filename, mediaType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	switch mediaType {
	case "image":
		return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" || ext == ".tiff"
	case "video":
		return ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".mov" || ext == ".wmv" || ext == ".flv"
	case "audio":
		return ext == ".wav" || ext == ".mp3" || ext == ".flac" || ext == ".aac" || ext == ".ogg"
	}

	return false
}

// isValidMessageFormat validates secret message file format
func isValidMessageFormat(filename, messageType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	switch messageType {
	case "image":
		return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" || ext == ".gif" || ext == ".tiff"
	case "video":
		return ext == ".mp4" || ext == ".avi" || ext == ".mkv" || ext == ".mov" || ext == ".wmv" || ext == ".flv" || ext == ".webm"
	case "audio":
		return ext == ".wav" || ext == ".mp3" || ext == ".flac" || ext == ".aac" || ext == ".ogg" || ext == ".m4a"
	}

	return false
}

// processEmbed handles the complete embedding process
func processEmbed(req *EmbedRequest) ([]byte, string, string, error) {
	// Create message data structure
	messageData := createMessageData(req)

	// Serialize to JSON
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return nil, "", "", errors.New("failed to serialize message data")
	}

	// Generate random salt for encryption
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, "", "", errors.New("failed to generate encryption salt")
	}

	// Generate encryption key from passphrase and salt
	key := utils.GenerateKey(req.Passphrase, salt)

	// Encrypt the message data
	encrypted, err := utils.EncryptData(jsonData, key)
	if err != nil {
		return nil, "", "", errors.New("failed to encrypt message data")
	}

	// Combine salt + encrypted data (salt is needed for decryption)
	fullData := append(salt, encrypted...)

	// Embed into carrier media based on type
	var result []byte
	var contentType, filename string

	switch req.MediaType {
	case "image":
		result, err = utils.EmbedDataInImage(req.Image, fullData)
		if err != nil {
			return nil, "", "", errors.New("failed to embed data in image: " + err.Error())
		}
		contentType = "image/png"
		filename = generateFilename(req.OriginalFilename, "embedded", ".png")

	case "video":
		result, err = utils.EmbedDataInVideo(req.VideoData, fullData)
		if err != nil {
			return nil, "", "", errors.New("failed to embed data in video: " + err.Error())
		}
		contentType = getVideoContentType(req.OriginalFilename)
		filename = generateFilename(req.OriginalFilename, "embedded", "")

	case "audio":
		result, err = utils.EmbedDataInAudio(req.AudioData, fullData)
		if err != nil {
			return nil, "", "", errors.New("failed to embed data in audio: " + err.Error())
		}
		contentType = getAudioContentType(req.OriginalFilename)
		filename = generateFilename(req.OriginalFilename, "embedded", "")
	}

	return result, contentType, filename, nil
}

// createMessageData creates the message data structure
func createMessageData(req *EmbedRequest) map[string]interface{} {
	messageData := map[string]interface{}{
		"type":      req.MessageType,
		"timestamp": getCurrentTimestamp(),
	}

	switch req.MessageType {
	case "text":
		messageData["content"] = req.Text
	case "audio":
		messageData["content"] = base64.StdEncoding.EncodeToString(req.MessageAudio)
		messageData["size"] = len(req.MessageAudio)
	case "image":
		messageData["content"] = base64.StdEncoding.EncodeToString(req.MessageImage)
		messageData["size"] = len(req.MessageImage)
	case "video":
		messageData["content"] = base64.StdEncoding.EncodeToString(req.MessageVideo)
		messageData["size"] = len(req.MessageVideo)
	}

	return messageData
}

// generateFilename generates output filename
func generateFilename(original, prefix, newExt string) string {
	if original == "" {
		if newExt != "" {
			return prefix + newExt
		}
		return prefix + "_file"
	}

	name := strings.TrimSuffix(original, filepath.Ext(original))
	if newExt != "" {
		return prefix + "_" + name + newExt
	}
	return prefix + "_" + original
}

// getVideoContentType returns content type for video files
func getVideoContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	default:
		return "video/mp4"
	}
}

// getAudioContentType returns content type for audio files
func getAudioContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	default:
		return "audio/wav"
	}
}

// getCurrentTimestamp returns current timestamp (you can implement this)
func getCurrentTimestamp() int64 {
	// Implementation depends on your time package
	// return time.Now().Unix()
	return 0 // placeholder
}

// respondError helper function for error responses
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, EmbedResponse{
		Success: false,
		Message: message,
	})
}
