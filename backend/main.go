package main

import (
	"fmt"
	"mangahub/backend/routes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Config stores application configuration
type Config struct {
	Port         string
	MangaRootDir string
	LogFile      string
}

// In a real application, you might load this from a file or environment variables
func loadConfig() Config {
	return Config{
		Port:         "8080",
		MangaRootDir: "../manga",
		LogFile:      "./manga-server.log",
	}
}

// We'll use a package-level logger for convenience
var zapLogger *zap.Logger

// setupZapLogger initializes the Zap logger
func setupZapLogger(config Config) {
	// For production, you could use zap.NewProduction() instead
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("Failed to initialize Zap logger: " + err.Error())
	}
	zapLogger = logger
}

// setupStaticDirs configures static file serving, including the "manga-images" folder
func setupStaticDirs(config Config, router *gin.Engine) {
	// Ensure manga directory exists
	if _, err := os.Stat(config.MangaRootDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.MangaRootDir, 0755)
		if err != nil {
			zapLogger.Fatal("Failed to create manga directory",
				zap.String("directory", config.MangaRootDir),
				zap.Error(err))
		}
	}

	// Serve manga images
	router.Static("/manga-images", config.MangaRootDir)

	// First build the frontend if you haven't already:
	// cd frontend && npm run build

	// Serve static files from the assets directory if it exists
	assetsPath := "./static/assets"
	if _, err := os.Stat(assetsPath); err == nil {
		router.Static("/assets", assetsPath)
	}

	// Serve individual static files with proper MIME types
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API and manga-images routes
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/manga-images") {
			c.Status(http.StatusNotFound)
			return
		}

		filePath := "./static" + path
		if _, err := os.Stat(filePath); err == nil {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".js" || ext == ".mjs" {
				c.Header("Content-Type", "application/javascript")
			} else if ext == ".css" {
				c.Header("Content-Type", "text/css")
			}
			c.File(filePath)
			return
		}

		// Default to index.html for SPA routing
		c.File("./static/index.html")
	})
}

func main() {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	config := loadConfig()

	// Initialize Zap logger
	setupZapLogger(config)
	defer zapLogger.Sync()

	router := gin.New()
	router.Use(gin.Recovery())

	// Custom logger middleware
	router.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()

		zapLogger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("clientIP", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", endTime.Sub(startTime)),
		)
	})

	// Setup static directories and routes
	setupStaticDirs(config, router)

	// Setup API routes
	routes.InitRoutes(config.MangaRootDir)
	routes.SetupRoutes(router)

	serverAddr := fmt.Sprintf(":%s", config.Port)
	zapLogger.Info("Starting manga server",
		zap.String("address", serverAddr),
	)

	if err := router.Run(serverAddr); err != nil {
		zapLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
