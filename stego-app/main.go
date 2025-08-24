package main

import (
	"net/http"
	"stego-app/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// API routes
	r.POST("/api/embed", handlers.EmbedHandler)
	r.POST("/api/extract", handlers.ExtractHandler)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve static files (frontend)
	r.Static("/public", "./public") // dùng /public thay vì /*filepath
	r.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html") // fallback
	})

	r.Run(":8080")
}
