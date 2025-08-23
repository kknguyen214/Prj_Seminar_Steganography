package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"io"
	"mime/multipart"
	"net/http"

	"stego-app/models"
	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

type EmbedRequest struct {
	Image       image.Image
	Passphrase  string
	MessageType string
	Text        string
	AudioData   []byte
	ImageData   []byte
}

func EmbedHandler(c *gin.Context) {
	// Parse request
	req, err := parseEmbedRequest(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Process embedding
	result, err := processEmbed(req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return result
	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", "attachment; filename=embedded_image.png")
	c.Data(http.StatusOK, "image/png", result)
}

// parseEmbedRequest parse toàn bộ request data
func parseEmbedRequest(c *gin.Context) (*EmbedRequest, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("failed to parse form")
	}

	// Parse image
	img, err := parseImage(form)
	if err != nil {
		return nil, err
	}

	req := &EmbedRequest{
		Image:       img,
		Passphrase:  c.PostForm("passphrase"),
		MessageType: c.PostForm("message_type"),
		Text:        c.PostForm("text"),
	}

	// Validate required fields
	if req.Passphrase == "" || req.MessageType == "" {
		return nil, errors.New("passphrase and message_type required")
	}

	// Parse content theo type
	if err := parseMessageContent(form, req); err != nil {
		return nil, err
	}

	return req, nil
}

// parseImage xử lý image upload
func parseImage(form *multipart.Form) (image.Image, error) {
	files := form.File["image"]
	if len(files) == 0 {
		return nil, errors.New("no image provided")
	}

	file := files[0]
	if utils.GetImageFormat(file.Filename) == "" {
		return nil, errors.New("unsupported image format")
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.New("invalid image format")
	}

	return img, nil
}

// parseMessageContent parse content theo message type
func parseMessageContent(form *multipart.Form, req *EmbedRequest) error {
	switch req.MessageType {
	case "audio":
		audioFiles := form.File["audio"]
		if len(audioFiles) == 0 {
			return errors.New("audio file required")
		}

		src, err := audioFiles[0].Open()
		if err != nil {
			return err
		}
		defer src.Close()

		req.AudioData, err = io.ReadAll(src)
		return err

	case "image":
		imgFiles := form.File["message_image"]
		if len(imgFiles) == 0 {
			return errors.New("message image required")
		}

		src, err := imgFiles[0].Open()
		if err != nil {
			return err
		}
		defer src.Close()

		req.ImageData, err = io.ReadAll(src)
		return err

	case "text":
		if req.Text == "" {
			return errors.New("text content required")
		}
		return nil

	default:
		return errors.New("invalid message_type. Must be: text, audio, or image")
	}
}

// processEmbed xử lý toàn bộ quy trình embed
func processEmbed(req *EmbedRequest) ([]byte, error) {
	// Tạo message data
	messageData := createMessageData(req)

	// Serialize to JSON
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return nil, err
	}

	// Generate salt và encrypt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := utils.GenerateKey(req.Passphrase, salt)
	encrypted, err := utils.EncryptData(jsonData, key)
	if err != nil {
		return nil, err
	}

	// Combine salt + encrypted data
	fullData := append(salt, encrypted...)

	// Embed vào image
	buf, err := utils.EmbedDataInImage(req.Image, fullData)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// createMessageData tạo message data structure
func createMessageData(req *EmbedRequest) map[string]interface{} {
	messageData := map[string]interface{}{
		"type": req.MessageType,
	}

	switch req.MessageType {
	case "text":
		messageData["content"] = req.Text
	case "audio":
		messageData["content"] = base64.StdEncoding.EncodeToString(req.AudioData)
	case "image":
		messageData["content"] = base64.StdEncoding.EncodeToString(req.ImageData)
	}

	return messageData
}

// respondError helper function cho error response
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.Response{
		Success: false,
		Message: message,
	})
}
