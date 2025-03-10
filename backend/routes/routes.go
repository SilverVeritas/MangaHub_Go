package routes

import (
	"mangahub/backend/models"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Global metadata manager
var metadataManager *models.MetadataManager

// InitRoutes initializes the routes with the given manga root directory
func InitRoutes(mangaRootDir string) {
	metadataManager = models.NewMetadataManager(mangaRootDir)
}

// SetupRoutes configures all the API routes for the manga reader
func SetupRoutes(router *gin.Engine) {
	// API Group
	api := router.Group("/api")
	{
		// Manga routes
		api.GET("/manga", listManga)
		api.GET("/manga/:id", getManga)
		api.GET("/manga/:id/chapters", listChapters)

		// Chapter routes
		api.GET("/manga/:id/chapter/:chapterNumber", getChapter)
		api.GET("/manga/:id/chapter/:chapterNumber/page/:pageNumber", getPage)

		// Search routes
		api.GET("/search", searchManga)

		// Optional: Admin routes for managing manga
		admin := api.Group("/admin")
		{
			admin.POST("/manga", addManga)
			admin.PUT("/manga/:id", updateManga)
			admin.POST("/manga/:id/chapter", addChapter)
			admin.PUT("/manga/:id/chapter/:chapterNumber", updateChapter)
		}
	}
}

// Route handlers

// listManga returns a list of all available manga series
func listManga(c *gin.Context) {
	mangas, err := metadataManager.ScanForManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga list: " + err.Error()})
		return
	}

	// Convert to response format
	var response []gin.H
	for _, manga := range mangas {
		response = append(response, gin.H{
			"id":           manga.ID,
			"title":        manga.Title,
			"description":  manga.Description,
			"coverImage":   manga.GetCoverImageURL(),
			"genres":       manga.Genres,
			"author":       manga.Author,
			"status":       manga.Status,
			"chapterCount": manga.ChapterCount,
		})
	}

	c.JSON(http.StatusOK, response)
}

// getManga returns details about a specific manga
func getManga(c *gin.Context) {
	id := c.Param("id")

	manga, err := metadataManager.GetMangaByID(id)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	// Map to response
	response := gin.H{
		"id":            manga.ID,
		"title":         manga.Title,
		"description":   manga.Description,
		"coverImage":    manga.GetCoverImageURL(),
		"genres":        manga.Genres,
		"author":        manga.Author,
		"artist":        manga.Artist,
		"status":        manga.Status,
		"publishedYear": manga.PublishedYear,
		"lastUpdated":   manga.LastUpdated,
		"chapterCount":  manga.ChapterCount,
		"altTitles":     manga.AltTitles,
	}

	c.JSON(http.StatusOK, response)
}

// listChapters returns a list of chapters for a specific manga
func listChapters(c *gin.Context) {
	mangaID := c.Param("id")

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	// Sort chapters by number (implement in models package or here)
	// For simplicity, I'll do it here, but it could be moved to a model method

	// Convert to response format
	var response []gin.H
	for _, chapter := range chapters {
		response = append(response, gin.H{
			"id":          chapter.ID,
			"mangaId":     chapter.MangaID,
			"number":      chapter.Number,
			"title":       chapter.Title,
			"releaseDate": chapter.ReleaseDate,
			"pageCount":   chapter.PageCount,
			"volume":      chapter.Volume,
			"special":     chapter.Special,
		})
	}

	c.JSON(http.StatusOK, response)
}

// getChapter returns details about a specific chapter
func getChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	// Find the requested chapter
	var targetChapter *models.Chapter
	for i := range chapters {
		if chapters[i].Number == chapterNumber {
			targetChapter = &chapters[i]
			break
		}
	}

	if targetChapter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Get pages for this chapter
	pages, err := targetChapter.GetPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pages: " + err.Error()})
		return
	}

	// Convert to response format
	response := gin.H{
		"id":          targetChapter.ID,
		"mangaId":     targetChapter.MangaID,
		"number":      targetChapter.Number,
		"title":       targetChapter.Title,
		"releaseDate": targetChapter.ReleaseDate,
		"pageCount":   targetChapter.PageCount,
		"volume":      targetChapter.Volume,
		"special":     targetChapter.Special,
		"pages":       []gin.H{},
	}

	// Add page info
	pagesList := []gin.H{}
	for _, page := range pages {
		pagesList = append(pagesList, gin.H{
			"number":   page.Number,
			"imageUrl": page.GetImageURL(),
		})
	}
	response["pages"] = pagesList

	c.JSON(http.StatusOK, response)
}

// getPage returns a specific page from a chapter
func getPage(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")
	pageNumberStr := c.Param("pageNumber")

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	// Find the requested chapter
	var targetChapter *models.Chapter
	var chapterIndex int
	for i := range chapters {
		if chapters[i].Number == chapterNumber {
			targetChapter = &chapters[i]
			chapterIndex = i
			break
		}
	}

	if targetChapter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Get pages for this chapter
	pages, err := targetChapter.GetPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pages: " + err.Error()})
		return
	}

	// Find the requested page
	var targetPage *models.Page
	for i := range pages {
		if pages[i].Number == pageNumber {
			targetPage = &pages[i]
			break
		}
	}

	if targetPage == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	// Determine next and previous chapters
	var nextChapter, prevChapter string

	// If this is the last page of the current chapter
	if pageNumber >= len(pages) && chapterIndex < len(chapters)-1 {
		nextChapter = strconv.FormatFloat(chapters[chapterIndex+1].Number, 'f', -1, 64)
	}

	// If this is the first page of the current chapter
	if pageNumber == 1 && chapterIndex > 0 {
		prevChapter = strconv.FormatFloat(chapters[chapterIndex-1].Number, 'f', -1, 64)
	}

	// Create the response
	response := gin.H{
		"imageUrl":   targetPage.GetImageURL(),
		"pageNumber": targetPage.Number,
		"totalPages": len(pages),
		"chapterID":  targetChapter.ID,
		"mangaID":    mangaID,
		"nextPage":   targetPage.GetNextPageNumber(),
		"prevPage":   targetPage.GetPrevPageNumber(),
	}

	if nextChapter != "" {
		response["nextChapter"] = nextChapter
	}

	if prevChapter != "" {
		response["prevChapter"] = prevChapter
	}

	c.JSON(http.StatusOK, response)
}

// searchManga handles searching for manga by title or filtering by genres
func searchManga(c *gin.Context) {
	query := c.Query("q")
	genre := c.Query("genre")

	mangas, err := metadataManager.ScanForManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga list: " + err.Error()})
		return
	}

	// Filter manga based on search criteria
	var results []models.MangaSeries
	for _, manga := range mangas {
		// Filter by title if query is provided
		if query != "" {
			if !containsIgnoreCase(manga.Title, query) && !containsIgnoreCase(manga.Description, query) {
				// Check alt titles too
				found := false
				for _, altTitle := range manga.AltTitles {
					if containsIgnoreCase(altTitle, query) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		// Filter by genre if provided
		if genre != "" {
			found := false
			for _, g := range manga.Genres {
				if equalIgnoreCase(g, genre) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		results = append(results, manga)
	}

	// Convert to response format
	var response []gin.H
	for _, manga := range results {
		response = append(response, gin.H{
			"id":          manga.ID,
			"title":       manga.Title,
			"description": manga.Description,
			"coverImage":  manga.GetCoverImageURL(),
			"genres":      manga.Genres,
			"author":      manga.Author,
		})
	}

	c.JSON(http.StatusOK, response)
}

// Admin route handlers - these would be protected in a production environment

// addManga adds a new manga series
func addManga(c *gin.Context) {
	var requestManga struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Artist      string   `json:"artist"`
		Genres      []string `json:"genres"`
		Status      string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&requestManga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create a manga ID from the title
	id := createSlug(requestManga.Title)

	// Check if manga already exists
	if _, err := metadataManager.GetMangaByID(id); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Manga with this ID already exists"})
		return
	}

	// Create the manga directory
	mangaPath := filepath.Join(metadataManager.RootDir, id)
	if err := os.MkdirAll(mangaPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create manga directory: " + err.Error()})
		return
	}

	// Create the manga metadata
	manga := models.MangaSeries{
		ID:          id,
		Title:       requestManga.Title,
		Description: requestManga.Description,
		Author:      requestManga.Author,
		Artist:      requestManga.Artist,
		Genres:      requestManga.Genres,
		Status:      requestManga.Status,
		Path:        mangaPath,
	}

	// Save metadata
	metadataPath := filepath.Join(mangaPath, models.MetadataFileName)
	if err := manga.SaveToJSON(metadataPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga metadata: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          manga.ID,
		"title":       manga.Title,
		"description": manga.Description,
		"author":      manga.Author,
		"artist":      manga.Artist,
		"genres":      manga.Genres,
		"status":      manga.Status,
	})
}

// updateManga updates an existing manga series
func updateManga(c *gin.Context) {
	id := c.Param("id")

	var requestManga struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Artist      string   `json:"artist"`
		Genres      []string `json:"genres"`
		Status      string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&requestManga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get existing manga
	manga, err := metadataManager.GetMangaByID(id)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	// Update fields if provided
	if requestManga.Title != "" {
		manga.Title = requestManga.Title
	}
	if requestManga.Description != "" {
		manga.Description = requestManga.Description
	}
	if requestManga.Author != "" {
		manga.Author = requestManga.Author
	}
	if requestManga.Artist != "" {
		manga.Artist = requestManga.Artist
	}
	if len(requestManga.Genres) > 0 {
		manga.Genres = requestManga.Genres
	}
	if requestManga.Status != "" {
		manga.Status = requestManga.Status
	}

	// Save metadata
	metadataPath := filepath.Join(manga.Path, models.MetadataFileName)
	if err := manga.SaveToJSON(metadataPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga metadata: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          manga.ID,
		"title":       manga.Title,
		"description": manga.Description,
		"author":      manga.Author,
		"artist":      manga.Artist,
		"genres":      manga.Genres,
		"status":      manga.Status,
	})
}

// addChapter adds a new chapter to a manga series
func addChapter(c *gin.Context) {
	mangaID := c.Param("id")

	var requestChapter struct {
		Number  float64 `json:"number" binding:"required"`
		Title   string  `json:"title"`
		Volume  int     `json:"volume"`
		Special bool    `json:"special"`
	}

	if err := c.ShouldBindJSON(&requestChapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get existing manga
	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	// Create a chapter ID from the number
	chapterID := "chapter-" + strconv.FormatFloat(requestChapter.Number, 'f', 1, 64)
	chapterID = createSlug(chapterID)

	// Create the chapter directory
	chapterPath := filepath.Join(manga.Path, chapterID)
	if err := os.MkdirAll(chapterPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chapter directory: " + err.Error()})
		return
	}

	// Create the chapter metadata
	chapter := models.Chapter{
		ID:          chapterID,
		MangaID:     mangaID,
		Number:      requestChapter.Number,
		Title:       requestChapter.Title,
		ReleaseDate: timeNow(),
		Path:        chapterPath,
		Volume:      requestChapter.Volume,
		Special:     requestChapter.Special,
	}

	// Save metadata
	metadataPath := filepath.Join(chapterPath, models.MetadataFileName)
	if err := chapter.SaveToJSON(metadataPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chapter metadata: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          chapter.ID,
		"mangaId":     chapter.MangaID,
		"number":      chapter.Number,
		"title":       chapter.Title,
		"releaseDate": chapter.ReleaseDate,
		"volume":      chapter.Volume,
		"special":     chapter.Special,
	})
}

// updateChapter updates an existing chapter
func updateChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	var requestChapter struct {
		Title   string `json:"title"`
		Volume  int    `json:"volume"`
		Special bool   `json:"special"`
	}

	if err := c.ShouldBindJSON(&requestChapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get existing manga
	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	// Get chapters
	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	// Find the chapter to update
	var targetChapter *models.Chapter
	for i := range chapters {
		if chapters[i].Number == chapterNumber {
			targetChapter = &chapters[i]
			break
		}
	}

	if targetChapter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	// Update fields if provided
	if requestChapter.Title != "" {
		targetChapter.Title = requestChapter.Title
	}
	targetChapter.Volume = requestChapter.Volume
	targetChapter.Special = requestChapter.Special

	// Save metadata
	metadataPath := filepath.Join(targetChapter.Path, models.MetadataFileName)
	if err := targetChapter.SaveToJSON(metadataPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chapter metadata: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          targetChapter.ID,
		"mangaId":     targetChapter.MangaID,
		"number":      targetChapter.Number,
		"title":       targetChapter.Title,
		"releaseDate": targetChapter.ReleaseDate,
		"volume":      targetChapter.Volume,
		"special":     targetChapter.Special,
	})
}

// Helper functions

// containsIgnoreCase checks if a string contains a substring, ignoring case
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// equalIgnoreCase checks if two strings are equal, ignoring case
func equalIgnoreCase(s1, s2 string) bool {
	return strings.ToLower(s1) == strings.ToLower(s2)
}

// createSlug creates a URL-friendly slug from a string
func createSlug(s string) string {
	// Replace spaces and special characters with hyphens
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters (except hyphens)
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")
	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}

// timeNow returns the current time
func timeNow() time.Time {
	return time.Now()
}
