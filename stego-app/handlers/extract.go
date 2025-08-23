package handlers

import (
	"encoding/json"
	"image"
	"net/http"

	"stego-app/models"
	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

func ExtractHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "No image provided"})
		return
	}

	format := utils.GetImageFormat(file.Filename)
	if format == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Unsupported image format"})
		return
	}

	src, _ := file.Open()
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Invalid image format"})
		return
	}

	passphrase := c.PostForm("passphrase")
	if passphrase == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Passphrase required"})
		return
	}

	data, err := utils.ExtractDataFromImage(img)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: err.Error()})
		return
	}

	if len(data) < 16 {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Invalid embedded data"})
		return
	}

	salt := data[:16]
	encrypted := data[16:]
	key := utils.GenerateKey(passphrase, salt)
	decrypted, err := utils.DecryptData(encrypted, key)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Failed to decrypt data"})
		return
	}

	var message map[string]interface{}
	json.Unmarshal(decrypted, &message)

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data: models.ExtractedData{
			MessageType: message["type"].(string),
			Content:     message["content"].(string),
		},
	})
}
