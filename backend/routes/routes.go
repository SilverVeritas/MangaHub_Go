package routes

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MangaSeries represents a manga series with its metadata
type MangaSeries struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverImage  string   `json:"coverImage"`
	Genres      []string `json:"genres"`
}

// Chapter represents a manga chapter with its metadata
type Chapter struct {
	ID          string `json:"id"`
	MangaID     string `json:"mangaId"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	ReleaseDate string `json:"releaseDate"`
	PageCount   int    `json:"pageCount"`
}

// PageResponse represents the response for a manga page
type PageResponse struct {
	ImageURL    string `json:"imageUrl"`
	PageNumber  int    `json:"pageNumber"`
	TotalPages  int    `json:"totalPages"`
	ChapterID   string `json:"chapterId"`
	MangaID     string `json:"mangaId"`
	NextChapter string `json:"nextChapter,omitempty"`
	PrevChapter string `json:"prevChapter,omitempty"`
}

// setupRoutes configures all the API routes for the manga reader
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
	// TODO: Implement fetching manga list from filesystem or database
	// For now, return sample data
	manga := []MangaSeries{
		{
			ID:          "one-piece",
			Title:       "One Piece",
			Description: "The story follows the adventures of Monkey D. Luffy, a boy whose body gained the properties of rubber after unintentionally eating a Devil Fruit.",
			CoverImage:  "/manga-images/one-piece/cover.jpg",
			Genres:      []string{"Adventure", "Fantasy", "Action"},
		},
		{
			ID:          "naruto",
			Title:       "Naruto",
			Description: "It tells the story of Naruto Uzumaki, a young ninja who seeks recognition from his peers and dreams of becoming the Hokage, the leader of his village.",
			CoverImage:  "/manga-images/naruto/cover.jpg",
			Genres:      []string{"Action", "Adventure", "Fantasy"},
		},
	}

	c.JSON(http.StatusOK, manga)
}

// getManga returns details about a specific manga
func getManga(c *gin.Context) {
	id := c.Param("id")

	// TODO: Implement fetching specific manga details
	// For now, return sample data
	manga := MangaSeries{
		ID:          id,
		Title:       "Sample Manga",
		Description: "This is a sample manga description.",
		CoverImage:  filepath.Join("/manga-images", id, "cover.jpg"),
		Genres:      []string{"Action", "Adventure"},
	}

	c.JSON(http.StatusOK, manga)
}

// listChapters returns a list of chapters for a specific manga
func listChapters(c *gin.Context) {
	mangaID := c.Param("id")

	// TODO: Implement fetching chapters from filesystem or database
	// For now, return sample data
	chapters := []Chapter{
		{
			ID:          "1",
			MangaID:     mangaID,
			Number:      1,
			Title:       "The Beginning",
			ReleaseDate: "2023-01-01",
			PageCount:   45,
		},
		{
			ID:          "2",
			MangaID:     mangaID,
			Number:      2,
			Title:       "The Journey Continues",
			ReleaseDate: "2023-01-15",
			PageCount:   38,
		},
	}

	c.JSON(http.StatusOK, chapters)
}

// getChapter returns details about a specific chapter
func getChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumber, err := strconv.Atoi(c.Param("chapterNumber"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	// TODO: Implement fetching chapter details
	// For now, return sample data
	chapter := Chapter{
		ID:          c.Param("chapterNumber"),
		MangaID:     mangaID,
		Number:      chapterNumber,
		Title:       "Sample Chapter",
		ReleaseDate: "2023-01-01",
		PageCount:   45,
	}

	c.JSON(http.StatusOK, chapter)
}

// getPage returns a specific page from a chapter
func getPage(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumber, err := strconv.Atoi(c.Param("chapterNumber"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chapter number"})
		return
	}

	pageNumber, err := strconv.Atoi(c.Param("pageNumber"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	// TODO: Implement fetching page details
	// For now, return sample data
	totalPages := 45

	pageResponse := PageResponse{
		ImageURL:   filepath.Join("/manga-images", mangaID, strconv.Itoa(chapterNumber), strconv.Itoa(pageNumber)+".jpg"),
		PageNumber: pageNumber,
		TotalPages: totalPages,
		ChapterID:  strconv.Itoa(chapterNumber),
		MangaID:    mangaID,
	}

	// Add next/previous chapter links if applicable
	if pageNumber >= totalPages {
		pageResponse.NextChapter = strconv.Itoa(chapterNumber + 1)
	}

	if chapterNumber > 1 && pageNumber == 1 {
		pageResponse.PrevChapter = strconv.Itoa(chapterNumber - 1)
	}

	c.JSON(http.StatusOK, pageResponse)
}

// searchManga handles searching for manga by title or filtering by genres
func searchManga(c *gin.Context) {
	query := c.Query("q")
	genre := c.Query("genre")

	// TODO: Implement actual search functionality
	// For now, return sample filtered data
	var results []MangaSeries

	if query != "" || genre != "" {
		results = []MangaSeries{
			{
				ID:          "sample-manga",
				Title:       "Sample Search Result",
				Description: "This is a sample search result.",
				CoverImage:  "/manga-images/sample-manga/cover.jpg",
				Genres:      []string{"Action", "Adventure"},
			},
		}
	} else {
		results = []MangaSeries{}
	}

	c.JSON(http.StatusOK, results)
}

// Admin route handlers - these would be protected in a production environment

// addManga adds a new manga series
func addManga(c *gin.Context) {
	var manga MangaSeries
	if err := c.ShouldBindJSON(&manga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement adding manga to filesystem or database

	c.JSON(http.StatusCreated, manga)
}

// updateManga updates an existing manga series
func updateManga(c *gin.Context) {
	id := c.Param("id")
	var manga MangaSeries
	if err := c.ShouldBindJSON(&manga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manga.ID = id

	// TODO: Implement updating manga in filesystem or database

	c.JSON(http.StatusOK, manga)
}

// addChapter adds a new chapter to a manga series
func addChapter(c *gin.Context) {
	mangaID := c.Param("id")
	var chapter Chapter
	if err := c.ShouldBindJSON(&chapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chapter.MangaID = mangaID

	// TODO: Implement adding chapter to filesystem or database

	c.JSON(http.StatusCreated, chapter)
}

// updateChapter updates an existing chapter
func updateChapter(c *gin.Context) {
	mangaID := c.Param("id")
	chapterNumber := c.Param("chapterNumber")
	var chapter Chapter
	if err := c.ShouldBindJSON(&chapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chapter.MangaID = mangaID
	chapter.ID = chapterNumber

	// TODO: Implement updating chapter in filesystem or database

	c.JSON(http.StatusOK, chapter)
}
