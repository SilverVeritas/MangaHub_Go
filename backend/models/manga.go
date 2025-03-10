package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

var mangaLogger *zap.Logger

func init() {
	l, _ := zap.NewDevelopment()
	mangaLogger = l
}

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
	Path          string    `json:"-"` // Internal use only
}

func (m *MangaSeries) Validate() error {
	mangaLogger.Debug("Validate called",
		zap.String("mangaID", m.ID),
		zap.String("title", m.Title),
	)
	if m.ID == "" {
		mangaLogger.Warn("Validation failed: ID is empty")
		return NewValidationError("manga ID is required")
	}
	if m.Title == "" {
		mangaLogger.Warn("Validation failed: title is empty", zap.String("mangaID", m.ID))
		return NewValidationError("manga title is required")
	}
	return nil
}

func (m *MangaSeries) LoadFromJSON(path string) error {
	mangaLogger.Info("LoadFromJSON called",
		zap.String("path", path),
	)

	file, err := os.ReadFile(path)
	if err != nil {
		mangaLogger.Error("Failed to read manga metadata file",
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to read manga metadata: " + err.Error())
	}

	if err := json.Unmarshal(file, m); err != nil {
		mangaLogger.Error("Failed to parse manga metadata",
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to parse manga metadata: " + err.Error())
	}

	m.Path = filepath.Dir(path)

	mangaLogger.Info("Manga metadata loaded",
		zap.String("mangaID", m.ID),
		zap.String("title", m.Title),
		zap.String("path", m.Path),
	)
	return nil
}

func (m *MangaSeries) SaveToJSON(path string) error {
	mangaLogger.Info("SaveToJSON called",
		zap.String("mangaID", m.ID),
		zap.String("path", path),
	)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		mangaLogger.Error("Failed to marshal manga metadata",
			zap.String("mangaID", m.ID),
			zap.Error(err),
		)
		return NewMetadataError("failed to marshal manga metadata: " + err.Error())
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		mangaLogger.Error("Failed to write manga metadata",
			zap.String("mangaID", m.ID),
			zap.String("path", path),
			zap.Error(err),
		)
		return NewMetadataError("failed to write manga metadata: " + err.Error())
	}

	mangaLogger.Info("Manga metadata saved", zap.String("mangaID", m.ID))
	return nil
}

func (m *MangaSeries) GetChaptersPath() string {
	mangaLogger.Debug("GetChaptersPath called",
		zap.String("mangaID", m.ID),
		zap.String("path", m.Path),
	)
	return m.Path
}

func (m *MangaSeries) GetCoverImagePath() string {
	fullPath := filepath.Join(m.Path, filepath.Base(m.CoverImage))
	mangaLogger.Debug("GetCoverImagePath called",
		zap.String("mangaID", m.ID),
		zap.String("coverImagePath", fullPath),
	)
	return fullPath
}

func (m *MangaSeries) GetCoverImageURL() string {
	url := "/manga-images/" + m.ID + "/" + filepath.Base(m.CoverImage)
	mangaLogger.Debug("GetCoverImageURL called",
		zap.String("mangaID", m.ID),
		zap.String("coverImageURL", url),
	)
	return url
}
