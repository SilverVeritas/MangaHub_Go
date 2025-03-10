package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// We'll use a package-level logger for convenience in this file.
// If you've already set up a logger elsewhere, you can replace this init with that.
var chapterLogger *zap.Logger

func init() {
	l, _ := zap.NewDevelopment()
	chapterLogger = l
}

// Chapter represents a manga chapter with its metadata
type Chapter struct {
	ID          string    `json:"id"`
	MangaID     string    `json:"mangaId"`
	Number      float64   `json:"number"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"releaseDate"`
	PageCount   int       `json:"pageCount"`
	Path        string    `json:"-"` // Internal use only, not exported to JSON
	Volume      int       `json:"volume,omitempty"`
	Special     bool      `json:"special,omitempty"`
}

// Validate checks if the chapter has all required fields
func (c *Chapter) Validate() error {
	chapterLogger.Debug("Validate called",
		zap.String("chapterID", c.ID),
		zap.Float64("chapterNumber", c.Number),
		zap.String("mangaID", c.MangaID),
	)
	if c.MangaID == "" {
		chapterLogger.Warn("Validation failed: mangaID is empty", zap.String("chapterID", c.ID))
		return NewValidationError("manga ID is required")
	}
	if c.Number <= 0 {
		chapterLogger.Warn("Validation failed: chapter number must be positive", zap.Float64("chapterNumber", c.Number))
		return NewValidationError("chapter number must be positive")
	}
	return nil
}

// LoadFromJSON loads chapter metadata from a JSON file
func (c *Chapter) LoadFromJSON(path string) error {
	chapterLogger.Info("LoadFromJSON called", zap.String("path", path))
	file, err := os.ReadFile(path)
	if err != nil {
		chapterLogger.Error("Failed to read chapter metadata file",
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to read chapter metadata: " + err.Error())
	}

	if err := json.Unmarshal(file, c); err != nil {
		chapterLogger.Error("Failed to parse chapter metadata",
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to parse chapter metadata: " + err.Error())
	}

	c.Path = filepath.Dir(path)

	chapterLogger.Info("Chapter metadata loaded",
		zap.String("chapterID", c.ID),
		zap.String("mangaID", c.MangaID),
		zap.Float64("chapterNumber", c.Number),
	)
	return nil
}

// SaveToJSON saves chapter metadata to a JSON file
func (c *Chapter) SaveToJSON(path string) error {
	chapterLogger.Info("SaveToJSON called",
		zap.String("chapterID", c.ID),
		zap.String("path", path),
	)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		chapterLogger.Error("Failed to marshal chapter metadata", zap.Error(err))
		return NewMetadataError("failed to marshal chapter metadata: " + err.Error())
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		chapterLogger.Error("Failed to write chapter metadata to file",
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to write chapter metadata: " + err.Error())
	}

	chapterLogger.Info("Chapter metadata saved", zap.String("chapterID", c.ID))
	return nil
}

// GetPages returns a slice of Page objects for this chapter
func (c *Chapter) GetPages() ([]Page, error) {
	chapterLogger.Info("GetPages called",
		zap.String("chapterID", c.ID),
		zap.String("mangaID", c.MangaID),
		zap.String("path", c.Path),
	)

	files, err := os.ReadDir(c.Path)
	if err != nil {
		chapterLogger.Error("Cannot read pages for chapter directory",
			zap.String("chapterPath", c.Path),
			zap.Error(err),
		)
		return nil, NewChapterNotFoundError(
			fmt.Sprintf("cannot read pages for chapter %v of manga %s", c.Number, c.MangaID))
	}

	var pages []Page
	for _, file := range files {
		if file.IsDir() || isMetadataFile(file.Name()) {
			continue
		}

		pageNumStr := filepath.Base(file.Name())
		pageNumStr = filepath.Ext(pageNumStr)
		pageNumStr = pageNumStr[:len(pageNumStr)-len(filepath.Ext(pageNumStr))]

		pageNum, convErr := strconv.Atoi(pageNumStr)
		if convErr != nil {
			pageNum = len(pages) + 1
		}

		page := Page{
			Number:    pageNum,
			ImagePath: filepath.Join(c.Path, file.Name()),
			ChapterID: c.ID,
			MangaID:   c.MangaID, // Make sure we set MangaID here
		}
		pages = append(pages, page)
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Number < pages[j].Number
	})

	c.PageCount = len(pages)

	chapterLogger.Info("Pages found",
		zap.String("chapterID", c.ID),
		zap.Int("pageCount", c.PageCount),
	)
	return pages, nil
}

// GetFirstPage returns the first page of the chapter
func (c *Chapter) GetFirstPage() (*Page, error) {
	chapterLogger.Info("GetFirstPage called", zap.String("chapterID", c.ID))
	pages, err := c.GetPages()
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		chapterLogger.Warn("No pages found in chapter", zap.String("chapterID", c.ID))
		return nil, NewPageNotFoundError("chapter has no pages")
	}
	chapterLogger.Info("First page retrieved", zap.Int("pageNumber", pages[0].Number))
	return &pages[0], nil
}

// GetPageByNumber returns a specific page by its number
func (c *Chapter) GetPageByNumber(pageNumber int) (*Page, error) {
	chapterLogger.Info("GetPageByNumber called",
		zap.String("chapterID", c.ID),
		zap.Int("requestedPageNumber", pageNumber),
	)
	pages, err := c.GetPages()
	if err != nil {
		return nil, err
	}
	for i, page := range pages {
		if page.Number == pageNumber {
			chapterLogger.Info("Page found", zap.Int("pageNumber", pageNumber))
			return &pages[i], nil
		}
	}
	chapterLogger.Warn("Requested page not found in chapter",
		zap.String("chapterID", c.ID),
		zap.Int("pageNumber", pageNumber),
	)
	return nil, NewPageNotFoundError(
		fmt.Sprintf("page %d not found in chapter %v", pageNumber, c.Number))
}

// Helper function to check if a file is a metadata file
func isMetadataFile(filename string) bool {
	return filename == "metadata.json" || filepath.Ext(filename) == ".json"
}
