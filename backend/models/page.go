package models

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"os"
	"path/filepath"
)

// Page represents a single page in a manga chapter
type Page struct {
	Number    int    `json:"number"`
	ImagePath string `json:"-"` // Internal use only, not exported to JSON
	ChapterID string `json:"chapterId"`
	MangaID   string `json:"mangaId"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	FileSize  int64  `json:"fileSize,omitempty"`
	MimeType  string `json:"mimeType,omitempty"`
}

// LoadImageMetadata loads image dimensions and other metadata
func (p *Page) LoadImageMetadata() error {
	// Get file info for size
	fileInfo, err := os.Stat(p.ImagePath)
	if err != nil {
		return NewMetadataError("failed to get page file info: " + err.Error())
	}
	p.FileSize = fileInfo.Size()

	// Open the image to get dimensions and type
	file, err := os.Open(p.ImagePath)
	if err != nil {
		return NewMetadataError("failed to open page image: " + err.Error())
	}
	defer file.Close()

	// Detect image format and dimensions
	img, format, err := image.DecodeConfig(file)
	if err != nil {
		return NewMetadataError("failed to decode page image: " + err.Error())
	}

	p.Width = img.Width
	p.Height = img.Height
	p.MimeType = "image/" + format

	return nil
}

// GetImageURL returns the URL for accessing this page
func (p *Page) GetImageURL() string {
	// Extract the relative path components we need
	dir := filepath.Dir(p.ImagePath)
	parts := filepath.SplitList(dir)

	// Construct a URL path suitable for your static file serving
	// This assumes your manga is organized as /manga/[manga-id]/[chapter-id]/[page].jpg
	mangaID := p.MangaID
	if mangaID == "" {
		// Try to extract from path if not set
		if len(parts) >= 2 {
			mangaID = parts[len(parts)-2]
		} else {
			mangaID = "unknown"
		}
	}

	chapterID := p.ChapterID
	if chapterID == "" {
		// Try to extract from path if not set
		if len(parts) >= 1 {
			chapterID = parts[len(parts)-1]
		} else {
			chapterID = "unknown"
		}
	}

	filename := filepath.Base(p.ImagePath)
	return fmt.Sprintf("/manga-images/%s/%s/%s", mangaID, chapterID, filename)
}

// Validate checks if the page has all required fields
func (p *Page) Validate() error {
	if p.Number <= 0 {
		return NewValidationError("page number must be positive")
	}
	if p.ChapterID == "" {
		return NewValidationError("chapter ID is required")
	}
	if p.ImagePath == "" {
		return NewValidationError("image path is required")
	}
	return nil
}

// ImageExists checks if the image file exists
func (p *Page) ImageExists() bool {
	_, err := os.Stat(p.ImagePath)
	return err == nil
}

// GetNextPageNumber returns the next page number or 0 if this is the last page
func (p *Page) GetNextPageNumber() int {
	return p.Number + 1
}

// GetPrevPageNumber returns the previous page number or 0 if this is the first page
func (p *Page) GetPrevPageNumber() int {
	if p.Number <= 1 {
		return 0
	}
	return p.Number - 1
}
