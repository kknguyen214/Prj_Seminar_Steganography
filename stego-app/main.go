package main

import (
	"log"
	"net/http"
	"os"
	"stego-app/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Cấu hình CORS cho production
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5500", "http://127.0.0.1:5500", "https://steganography-lab.onrender.com"}
	r.Use(cors.New(config))

	// API routes
	r.POST("/api/embed", handlers.EmbedHandler)
	r.POST("/api/extract", handlers.ExtractHandler)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Lấy port từ biến môi trường của Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Chạy port 8080 nếu không có biến môi trường (khi chạy local)
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	r.Run(":" + port)
}
