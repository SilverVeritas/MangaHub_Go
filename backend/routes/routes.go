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
	"go.uber.org/zap"
)

var (
	metadataManager *models.MetadataManager
	zapLogger       *zap.Logger
)

// init sets up zap logger for this package
func init() {
	l, _ := zap.NewDevelopment()
	zapLogger = l
}

// InitRoutes initializes the routes with the given manga root directory
func InitRoutes(mangaRootDir string) {
	zapLogger.Info("InitRoutes called", zap.String("mangaRootDir", mangaRootDir))
	metadataManager = models.NewMetadataManager(mangaRootDir)
}

// SetupRoutes configures all the API routes for the manga reader
func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/manga", listManga)
		api.GET("/manga/:id", getManga)
		api.GET("/manga/:id/chapters", listChapters)

		api.GET("/manga/:id/chapter/:chapterNumber", getChapter)
		api.GET("/manga/:id/chapter/:chapterNumber/page/:pageNumber", getPage)

		api.GET("/search", searchManga)

		admin := api.Group("/admin")
		{
			admin.POST("/manga", addManga)
			admin.PUT("/manga/:id", updateManga)
			admin.POST("/manga/:id/chapter", addChapter)
			admin.PUT("/manga/:id/chapter/:chapterNumber", updateChapter)
		}
	}
}

// listManga returns a list of all available manga series
func listManga(c *gin.Context) {
	zapLogger.Info("listManga handler called")

	mangas, err := metadataManager.ScanForManga()
	if err != nil {
		zapLogger.Error("Failed to retrieve manga list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga list: " + err.Error()})
		return
	}

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

	zapLogger.Info("listManga returning data", zap.Int("mangaCount", len(response)))
	c.JSON(http.StatusOK, response)
}

// getManga returns details about a specific manga
func getManga(c *gin.Context) {
	id := c.Param("id")
	zapLogger.Info("getManga handler called", zap.String("mangaID", id))

	manga, err := metadataManager.GetMangaByID(id)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

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

	zapLogger.Info("getManga returning data", zap.String("mangaID", manga.ID))
	c.JSON(http.StatusOK, response)
}

// listChapters returns a list of chapters for a specific manga
func listChapters(c *gin.Context) {
	mangaID := c.Param("id")
	zapLogger.Info("listChapters handler called", zap.String("mangaID", mangaID))

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", mangaID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		zapLogger.Error("Failed to retrieve chapters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

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

	zapLogger.Info("listChapters returning data", zap.Int("chapterCount", len(response)))
	c.JSON(http.StatusOK, response)
}

// getChapter returns details about a specific chapter
func getChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")
	zapLogger.Info("getChapter handler called",
		zap.String("mangaID", mangaID),
		zap.String("chapterNumber", chapterNumberStr),
	)

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		zapLogger.Warn("Invalid chapter number", zap.String("chapterNumberStr", chapterNumberStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", mangaID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		zapLogger.Error("Failed to retrieve chapters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	var targetChapter *models.Chapter
	for i := range chapters {
		if chapters[i].Number == chapterNumber {
			targetChapter = &chapters[i]
			break
		}
	}

	if targetChapter == nil {
		zapLogger.Warn("Chapter not found",
			zap.String("mangaID", mangaID),
			zap.Float64("chapterNumber", chapterNumber),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	pages, err := targetChapter.GetPages()
	if err != nil {
		zapLogger.Error("Failed to retrieve pages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pages: " + err.Error()})
		return
	}

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

	var pagesList []gin.H
	for _, page := range pages {
		pagesList = append(pagesList, gin.H{
			"number":   page.Number,
			"imageUrl": page.GetImageURL(),
		})
	}
	response["pages"] = pagesList

	zapLogger.Info("getChapter returning data", zap.String("chapterID", targetChapter.ID))
	c.JSON(http.StatusOK, response)
}

// getPage returns a specific page from a chapter
func getPage(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")
	pageNumberStr := c.Param("pageNumber")
	zapLogger.Info("getPage handler called",
		zap.String("mangaID", mangaID),
		zap.String("chapterNumber", chapterNumberStr),
		zap.String("pageNumber", pageNumberStr),
	)

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		zapLogger.Warn("Invalid chapter number", zap.String("chapterNumberStr", chapterNumberStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil {
		zapLogger.Warn("Invalid page number", zap.String("pageNumberStr", pageNumberStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", mangaID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		zapLogger.Error("Failed to retrieve chapters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

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
		zapLogger.Warn("Chapter not found",
			zap.String("mangaID", mangaID),
			zap.Float64("chapterNumber", chapterNumber),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	pages, err := targetChapter.GetPages()
	if err != nil {
		zapLogger.Error("Failed to retrieve pages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pages: " + err.Error()})
		return
	}

	var targetPage *models.Page
	for i := range pages {
		if pages[i].Number == pageNumber {
			targetPage = &pages[i]
			break
		}
	}

	if targetPage == nil {
		zapLogger.Warn("Page not found",
			zap.String("mangaID", mangaID),
			zap.Float64("chapterNumber", chapterNumber),
			zap.Int("pageNumber", pageNumber),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	var nextChapter, prevChapter string
	if pageNumber >= len(pages) && chapterIndex < len(chapters)-1 {
		nextChapter = strconv.FormatFloat(chapters[chapterIndex+1].Number, 'f', -1, 64)
	}
	if pageNumber == 1 && chapterIndex > 0 {
		prevChapter = strconv.FormatFloat(chapters[chapterIndex-1].Number, 'f', -1, 64)
	}

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

	zapLogger.Info("getPage returning data",
		zap.String("chapterID", targetChapter.ID),
		zap.Int("pageNumber", targetPage.Number),
	)
	c.JSON(http.StatusOK, response)
}

// searchManga handles searching for manga by title or filtering by genres
func searchManga(c *gin.Context) {
	query := c.Query("q")
	genre := c.Query("genre")

	zapLogger.Info("searchManga called",
		zap.String("query", query),
		zap.String("genre", genre),
	)

	mangas, err := metadataManager.ScanForManga()
	if err != nil {
		zapLogger.Error("Failed to retrieve manga list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga list: " + err.Error()})
		return
	}

	var results []models.MangaSeries
	for _, manga := range mangas {
		if query != "" {
			if !containsIgnoreCase(manga.Title, query) && !containsIgnoreCase(manga.Description, query) {
				foundAlt := false
				for _, altTitle := range manga.AltTitles {
					if containsIgnoreCase(altTitle, query) {
						foundAlt = true
						break
					}
				}
				if !foundAlt {
					continue
				}
			}
		}
		if genre != "" {
			foundGenre := false
			for _, g := range manga.Genres {
				if equalIgnoreCase(g, genre) {
					foundGenre = true
					break
				}
			}
			if !foundGenre {
				continue
			}
		}
		results = append(results, manga)
	}

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

	zapLogger.Info("searchManga returning results", zap.Int("resultsCount", len(response)))
	c.JSON(http.StatusOK, response)
}

func addManga(c *gin.Context) {
	zapLogger.Info("addManga handler called")

	var requestManga struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Artist      string   `json:"artist"`
		Genres      []string `json:"genres"`
		Status      string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&requestManga); err != nil {
		zapLogger.Warn("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	id := createSlug(requestManga.Title)
	if _, err := metadataManager.GetMangaByID(id); err == nil {
		zapLogger.Warn("Manga with this ID already exists", zap.String("id", id))
		c.JSON(http.StatusConflict, gin.H{"error": "Manga with this ID already exists"})
		return
	}

	mangaPath := filepath.Join(metadataManager.RootDir, id)
	if err := os.MkdirAll(mangaPath, 0755); err != nil {
		zapLogger.Error("Failed to create manga directory", zap.String("mangaPath", mangaPath), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create manga directory: " + err.Error()})
		return
	}

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

	metadataPath := filepath.Join(mangaPath, models.MetadataFileName)
	if err := manga.SaveToJSON(metadataPath); err != nil {
		zapLogger.Error("Failed to save manga metadata",
			zap.String("metadataPath", metadataPath),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga metadata: " + err.Error()})
		return
	}

	zapLogger.Info("Manga created", zap.String("mangaID", manga.ID))
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

func updateManga(c *gin.Context) {
	id := c.Param("id")
	zapLogger.Info("updateManga handler called", zap.String("mangaID", id))

	var requestManga struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Artist      string   `json:"artist"`
		Genres      []string `json:"genres"`
		Status      string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&requestManga); err != nil {
		zapLogger.Warn("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	manga, err := metadataManager.GetMangaByID(id)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

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

	metadataPath := filepath.Join(manga.Path, models.MetadataFileName)
	if err := manga.SaveToJSON(metadataPath); err != nil {
		zapLogger.Error("Failed to save manga metadata",
			zap.String("metadataPath", metadataPath),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save manga metadata: " + err.Error()})
		return
	}

	zapLogger.Info("Manga updated", zap.String("mangaID", manga.ID))
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

func addChapter(c *gin.Context) {
	mangaID := c.Param("id")
	zapLogger.Info("addChapter handler called", zap.String("mangaID", mangaID))

	var requestChapter struct {
		Number  float64 `json:"number" binding:"required"`
		Title   string  `json:"title"`
		Volume  int     `json:"volume"`
		Special bool    `json:"special"`
	}

	if err := c.ShouldBindJSON(&requestChapter); err != nil {
		zapLogger.Warn("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", mangaID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapterID := "chapter-" + strconv.FormatFloat(requestChapter.Number, 'f', 1, 64)
	chapterID = createSlug(chapterID)

	chapterPath := filepath.Join(manga.Path, chapterID)
	if err := os.MkdirAll(chapterPath, 0755); err != nil {
		zapLogger.Error("Failed to create chapter directory",
			zap.String("chapterPath", chapterPath),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chapter directory: " + err.Error()})
		return
	}

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

	metadataPath := filepath.Join(chapterPath, models.MetadataFileName)
	if err := chapter.SaveToJSON(metadataPath); err != nil {
		zapLogger.Error("Failed to save chapter metadata",
			zap.String("metadataPath", metadataPath),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chapter metadata: " + err.Error()})
		return
	}

	zapLogger.Info("Chapter created",
		zap.String("mangaID", mangaID),
		zap.String("chapterID", chapter.ID),
	)
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

func updateChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumberStr := c.Param("chapterNumber")
	zapLogger.Info("updateChapter handler called",
		zap.String("mangaID", mangaID),
		zap.String("chapterNumberStr", chapterNumberStr),
	)

	chapterNumber, err := strconv.ParseFloat(chapterNumberStr, 64)
	if err != nil {
		zapLogger.Warn("Invalid chapter number", zap.String("chapterNumberStr", chapterNumberStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	var requestChapter struct {
		Title   string `json:"title"`
		Volume  int    `json:"volume"`
		Special bool   `json:"special"`
	}

	if err := c.ShouldBindJSON(&requestChapter); err != nil {
		zapLogger.Warn("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	manga, err := metadataManager.GetMangaByID(mangaID)
	if err != nil {
		if models.IsMangaNotFoundError(err) {
			zapLogger.Warn("Manga not found", zap.String("mangaID", mangaID))
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			zapLogger.Error("Failed to retrieve manga", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve manga: " + err.Error()})
		}
		return
	}

	chapters, err := metadataManager.ScanForChapters(manga)
	if err != nil {
		zapLogger.Error("Failed to retrieve chapters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chapters: " + err.Error()})
		return
	}

	var targetChapter *models.Chapter
	for i := range chapters {
		if chapters[i].Number == chapterNumber {
			targetChapter = &chapters[i]
			break
		}
	}

	if targetChapter == nil {
		zapLogger.Warn("Chapter not found",
			zap.String("mangaID", mangaID),
			zap.Float64("chapterNumber", chapterNumber),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}

	if requestChapter.Title != "" {
		targetChapter.Title = requestChapter.Title
	}
	targetChapter.Volume = requestChapter.Volume
	targetChapter.Special = requestChapter.Special

	metadataPath := filepath.Join(targetChapter.Path, models.MetadataFileName)
	if err := targetChapter.SaveToJSON(metadataPath); err != nil {
		zapLogger.Error("Failed to save chapter metadata",
			zap.String("metadataPath", metadataPath),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chapter metadata: " + err.Error()})
		return
	}

	zapLogger.Info("Chapter updated",
		zap.String("mangaID", mangaID),
		zap.String("chapterID", targetChapter.ID),
	)
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

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func equalIgnoreCase(s1, s2 string) bool {
	return strings.ToLower(s1) == strings.ToLower(s2)
}

func createSlug(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	return slug
}

func timeNow() time.Time {
	return time.Now()
}
