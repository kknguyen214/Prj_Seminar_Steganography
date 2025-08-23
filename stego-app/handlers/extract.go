package handlers

import (
	"encoding/json"
	"errors"
	"image"
	"net/http"

	"stego-app/models"
	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

type ExtractRequest struct {
	Image      image.Image
	Passphrase string
}

func ExtractHandler(c *gin.Context) {
	// Parse request
	req, err := parseExtractRequest(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Process extraction
	result, err := processExtraction(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Return success response
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    result,
	})
}

// parseExtractRequest parse và validate request
func parseExtractRequest(c *gin.Context) (*ExtractRequest, error) {
	// Parse image file
	img, err := parseImageFromForm(c)
	if err != nil {
		return nil, err
	}

	// Get passphrase
	passphrase := c.PostForm("passphrase")
	if passphrase == "" {
		return nil, errors.New("passphrase required")
	}

	return &ExtractRequest{
		Image:      img,
		Passphrase: passphrase,
	}, nil
}

// parseImageFromForm parse image từ form data
func parseImageFromForm(c *gin.Context) (image.Image, error) {
	file, err := c.FormFile("image")
	if err != nil {
		return nil, errors.New("no image provided")
	}

	// Validate image format
	format := utils.GetImageFormat(file.Filename)
	if format == "" {
		return nil, errors.New("unsupported image format")
	}

	// Open and decode image
	src, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to open image file")
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.New("invalid image format")
	}

	return img, nil
}

// processExtraction xử lý toàn bộ quá trình extract
func processExtraction(req *ExtractRequest) (*models.ExtractedData, error) {
	// Extract raw data từ image
	rawData, err := extractRawData(req.Image)
	if err != nil {
		return nil, err
	}

	// Decrypt data
	decryptedData, err := decryptExtractedData(rawData, req.Passphrase)
	if err != nil {
		return nil, err
	}

	// Parse message data
	extractedData, err := parseExtractedMessage(decryptedData)
	if err != nil {
		return nil, err
	}

	return extractedData, nil
}

// extractRawData extract dữ liệu thô từ image
func extractRawData(img image.Image) ([]byte, error) {
	data, err := utils.ExtractDataFromImage(img)
	if err != nil {
		return nil, errors.New("failed to extract data from image")
	}

	if len(data) < 16 {
		return nil, errors.New("invalid embedded data")
	}

	return data, nil
}

// decryptExtractedData decrypt dữ liệu đã extract
func decryptExtractedData(data []byte, passphrase string) ([]byte, error) {
	// Tách salt và encrypted data
	salt := data[:16]
	encrypted := data[16:]

	// Generate key và decrypt
	key := utils.GenerateKey(passphrase, salt)
	decrypted, err := utils.DecryptData(encrypted, key)
	if err != nil {
		return nil, errors.New("failed to decrypt data - wrong passphrase?")
	}

	return decrypted, nil
}

// parseExtractedMessage parse JSON message thành struct
func parseExtractedMessage(data []byte) (*models.ExtractedData, error) {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, errors.New("invalid message format")
	}

	// Validate required fields
	messageType, ok := message["type"].(string)
	if !ok {
		return nil, errors.New("missing or invalid message type")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, errors.New("missing or invalid content")
	}

	return &models.ExtractedData{
		MessageType: messageType,
		Content:     content,
	}, nil
}
