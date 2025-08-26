package main

import (
	"log"
	"net/http"
	"stego-app/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Request từ trình duyệt đến Nginx, rồi Nginx gọi đến backend -> giao tiếp server-to-server.

	// API routes
	r.POST("/api/embed", handlers.EmbedHandler)
	r.POST("/api/extract", handlers.ExtractHandler)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Port được định nghĩa trong docker-compose, chạy cố định ở đây.
	port := "8080"
	log.Printf("API server listening on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
