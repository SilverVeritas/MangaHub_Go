package main

import (
	"fmt"
	"log"
	"mangahub/backend/routes"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// Config stores application configuration
type Config struct {
	Port         string
	MangaRootDir string
	LogFile      string
}

func loadConfig() Config {
	// In a real application, you might load this from a file or environment variables
	return Config{
		Port:         "8080",
		MangaRootDir: "./manga",
		LogFile:      "./manga-server.log",
	}
}

func setupLogger(config Config) *os.File {
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(logFile)
	return logFile
}

func setupStaticDirs(config Config, router *gin.Engine) {
	// Ensure manga directory exists
	if _, err := os.Stat(config.MangaRootDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.MangaRootDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create manga directory: %v", err)
		}
	}

	// Serve static files (manga images)
	router.Static("/manga-images", config.MangaRootDir)

	// Serve frontend static files
	router.Static("/static", "./static")

	// Serve the main HTML file for all routes not matched
	router.NoRoute(func(c *gin.Context) {
		// Only serve the index file for regular page requests, not for API or static files
		path := c.Request.URL.Path
		if filepath.Ext(path) == "" && !filepath.HasPrefix(path, "/api") {
			c.File("./static/index.html")
		} else {
			c.Status(http.StatusNotFound)
		}
	})
}

func main() {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	// Load application configuration
	config := loadConfig()

	// Setup logger
	logFile := setupLogger(config)
	defer logFile.Close()

	// Initialize Gin router
	router := gin.New()

	// Use Gin's Recovery middleware to recover from panics
	router.Use(gin.Recovery())

	// Custom logger middleware
	router.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		log.Printf(
			"[%s] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			endTime.Sub(startTime),
		)
	})

	// Setup static directories and routes
	setupStaticDirs(config, router)

	// Setup API routes
	routes.SetupRoutes(router)

	// Start the server
	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Starting manga server on http://localhost%s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
