package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"stego-app/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add middleware for request logging
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Recovery middleware
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "steganography-api",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Steganography endpoints
		api.POST("/embed", handlers.EmbedHandler)
		api.POST("/extract", handlers.ExtractHandler)

		// Info endpoint
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service":     "Steganography API",
				"version":     "1.0.0",
				"description": "Hide and extract secret messages in images, audio, and video files",
				"endpoints": gin.H{
					"embed":   "POST /api/v1/embed - Hide secret message in media file",
					"extract": "POST /api/v1/extract - Extract secret message from media file",
				},
				"supported_carrier_formats": gin.H{
					"image": []string{"PNG", "JPG", "JPEG", "BMP", "TIFF"},
					"audio": []string{"WAV", "MP3", "FLAC", "AAC", "OGG"},
					"video": []string{"MP4", "AVI", "MKV", "MOV", "WMV", "FLV"},
				},
				"supported_message_formats": gin.H{
					"text":  "Plain text messages",
					"image": []string{"PNG", "JPG", "JPEG", "BMP", "GIF", "TIFF"},
					"audio": []string{"WAV", "MP3", "FLAC", "AAC", "OGG", "M4A"},
					"video": []string{"MP4", "AVI", "MKV", "MOV", "WMV", "FLV", "WEBM"},
				},
			})
		})
	}

	// Serve static files (optional - for web interface)
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Web interface (optional)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Steganography Tool",
		})
	})

	// Start server
	port := ":8080"
	log.Printf("🚀 Steganography API server starting on port %s", port)
	log.Printf("📖 API Documentation: http://localhost%s/api/v1/info", port)
	log.Printf("🌐 Web Interface: http://localhost%s", port)

	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
