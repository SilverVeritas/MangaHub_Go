package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// MangaSeries represents a manga series with all its metadata
type MangaSeries struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Author        string    `json:"author"`
	Artist        string    `json:"artist,omitempty"`
	CoverImage    string    `json:"coverImage"`
	Genres        []string  `json:"genres"`
	Status        string    `json:"status"`
	PublishedYear int       `json:"publishedYear,omitempty"`
	LastUpdated   time.Time `json:"lastUpdated"`
	ChapterCount  int       `json:"chapterCount"`
	AltTitles     []string  `json:"altTitles,omitempty"`
	Path          string    `json:"-"` // Internal use only, not exported to JSON
}

// Validate checks if the manga has all required fields
func (m *MangaSeries) Validate() error {
	if m.ID == "" {
		return NewValidationError("manga ID is required")
	}
	if m.Title == "" {
		return NewValidationError("manga title is required")
	}
	return nil
}

// LoadFromJSON loads manga metadata from a JSON file
func (m *MangaSeries) LoadFromJSON(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return NewMetadataError("failed to read manga metadata: " + err.Error())
	}

	if err := json.Unmarshal(file, m); err != nil {
		return NewMetadataError("failed to parse manga metadata: " + err.Error())
	}

	// Set the filesystem path
	m.Path = filepath.Dir(path)

	return nil
}

// SaveToJSON saves manga metadata to a JSON file
func (m *MangaSeries) SaveToJSON(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return NewMetadataError("failed to marshal manga metadata: " + err.Error())
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return NewMetadataError("failed to write manga metadata: " + err.Error())
	}

	return nil
}

// GetChaptersPath returns the path where chapters for this manga are stored
func (m *MangaSeries) GetChaptersPath() string {
	return m.Path
}

// GetCoverImagePath returns the absolute path to the cover image
func (m *MangaSeries) GetCoverImagePath() string {
	return filepath.Join(m.Path, filepath.Base(m.CoverImage))
}

// GetCoverImageURL returns the URL for the cover image
func (m *MangaSeries) GetCoverImageURL() string {
	// This will need to be adjusted based on your static file serving setup
	return "/manga-images/" + m.ID + "/" + filepath.Base(m.CoverImage)
}
