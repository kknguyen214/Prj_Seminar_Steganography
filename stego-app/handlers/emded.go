package handlers

import (
	"encoding/base64"
	"encoding/json"
	"image"
	"io"
	"math/rand"
	"net/http"

	"stego-app/models"
	"stego-app/utils"

	"github.com/gin-gonic/gin"
)

func EmbedHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Failed to parse form"})
		return
	}

	files := form.File["image"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "No image provided"})
		return
	}
	file := files[0]

	format := utils.GetImageFormat(file.Filename)
	if format == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Unsupported image format",
		})
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
	messageType := c.PostForm("message_type")
	text := c.PostForm("text")

	if passphrase == "" || messageType == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Passphrase and message_type required"})
		return
	}

	messageData := map[string]interface{}{"type": messageType, "content": text}

	switch messageType {
	case "audio":
		audioFiles := form.File["audio"]
		if len(audioFiles) == 0 {
			c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Audio file required"})
			return
		}
		audioSrc, _ := audioFiles[0].Open()
		defer audioSrc.Close()
		data, _ := io.ReadAll(audioSrc)
		messageData["content"] = base64.StdEncoding.EncodeToString(data)
	case "image":
		imgFiles := form.File["message_image"]
		if len(imgFiles) == 0 {
			c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: "Message image required"})
			return
		}
		imgSrc, _ := imgFiles[0].Open()
		defer imgSrc.Close()
		data, _ := io.ReadAll(imgSrc)
		messageData["content"] = base64.StdEncoding.EncodeToString(data)
	}

	jsonData, _ := json.Marshal(messageData)
	salt := make([]byte, 16)
	rand.Read(salt)
	key := utils.GenerateKey(passphrase, salt)
	encrypted, _ := utils.EncryptData(jsonData, key)
	fullData := append(salt, encrypted...)

	buf, err := utils.EmbedDataInImage(img, fullData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Success: false, Message: err.Error()})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", "attachment; filename=embedded_image.png")
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}
