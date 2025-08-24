package main

import (
	"net/http"
	"stego-app/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Allow all CORS
	r.Use(cors.Default())

	// API routes
	r.POST("/api/embed", handlers.EmbedHandler)
	r.POST("/api/extract", handlers.ExtractHandler)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
