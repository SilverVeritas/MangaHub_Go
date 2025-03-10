package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Chapter represents a manga chapter with its metadata
type Chapter struct {
	ID          string    `json:"id"`
	MangaID     string    `json:"mangaId"`
	Number      float64   `json:"number"` // Using float to support chapters like 1.5
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"releaseDate"`
	PageCount   int       `json:"pageCount"`
	Path        string    `json:"-"` // Internal use only, not exported to JSON
	Volume      int       `json:"volume,omitempty"`
	Special     bool      `json:"special,omitempty"`
}

// Validate checks if the chapter has all required fields
func (c *Chapter) Validate() error {
	if c.MangaID == "" {
		return NewValidationError("manga ID is required")
	}
	if c.Number <= 0 {
		return NewValidationError("chapter number must be positive")
	}
	return nil
}

// LoadFromJSON loads chapter metadata from a JSON file
func (c *Chapter) LoadFromJSON(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return NewMetadataError("failed to read chapter metadata: " + err.Error())
	}

	if err := json.Unmarshal(file, c); err != nil {
		return NewMetadataError("failed to parse chapter metadata: " + err.Error())
	}

	// Set the filesystem path
	c.Path = filepath.Dir(path)

	return nil
}

// SaveToJSON saves chapter metadata to a JSON file
func (c *Chapter) SaveToJSON(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return NewMetadataError("failed to marshal chapter metadata: " + err.Error())
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return NewMetadataError("failed to write chapter metadata: " + err.Error())
	}

	return nil
}

// GetPages returns a slice of Page objects for this chapter
func (c *Chapter) GetPages() ([]Page, error) {
	files, err := os.ReadDir(c.Path)
	if err != nil {
		return nil, NewChapterNotFoundError(
			fmt.Sprintf("cannot read pages for chapter %v of manga %s", c.Number, c.MangaID))
	}

	var pages []Page
	for _, file := range files {
		// Skip directories and non-image files
		if file.IsDir() || isMetadataFile(file.Name()) {
			continue
		}

		// Extract page number from filename (assuming filenames like 1.jpg, 2.png, etc.)
		pageNumStr := filepath.Base(file.Name())
		pageNumStr = filepath.Ext(pageNumStr)
		pageNumStr = pageNumStr[:len(pageNumStr)-len(filepath.Ext(pageNumStr))]

		pageNum, err := strconv.Atoi(pageNumStr)
		if err != nil {
			// If we can't parse the page number, try to use position in the slice
			pageNum = len(pages) + 1
		}

		page := Page{
			Number:    pageNum,
			ImagePath: filepath.Join(c.Path, file.Name()),
			ChapterID: c.ID,
		}
		pages = append(pages, page)
	}

	// Sort pages by number
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Number < pages[j].Number
	})

	// Update page count
	c.PageCount = len(pages)

	return pages, nil
}

// GetFirstPage returns the first page of the chapter
func (c *Chapter) GetFirstPage() (*Page, error) {
	pages, err := c.GetPages()
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, NewPageNotFoundError("chapter has no pages")
	}

	return &pages[0], nil
}

// GetPageByNumber returns a specific page by its number
func (c *Chapter) GetPageByNumber(pageNumber int) (*Page, error) {
	pages, err := c.GetPages()
	if err != nil {
		return nil, err
	}

	for i, page := range pages {
		if page.Number == pageNumber {
			return &pages[i], nil
		}
	}

	return nil, NewPageNotFoundError(
		fmt.Sprintf("page %d not found in chapter %v", pageNumber, c.Number))
}

// Helper function to check if a file is a metadata file
func isMetadataFile(filename string) bool {
	return filename == "metadata.json" || filepath.Ext(filename) == ".json"
}
