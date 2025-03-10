package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MetadataFileName is the standard name for metadata files
	MetadataFileName = "metadata.json"
)

// MetadataManager provides utilities for managing metadata
type MetadataManager struct {
	RootDir string // Root directory for manga storage
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(rootDir string) *MetadataManager {
	return &MetadataManager{
		RootDir: rootDir,
	}
}

// ScanForManga scans the root directory for manga series
func (mm *MetadataManager) ScanForManga() ([]MangaSeries, error) {
	var mangas []MangaSeries

	// Read the root directory
	dirs, err := os.ReadDir(mm.RootDir)
	if err != nil {
		return nil, NewMetadataError("failed to read root directory: " + err.Error())
	}

	// Look for manga directories
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		// Check for metadata.json
		mangaPath := filepath.Join(mm.RootDir, dir.Name())
		metadataPath := filepath.Join(mangaPath, MetadataFileName)

		// If metadata exists, load it
		if _, err := os.Stat(metadataPath); err == nil {
			var manga MangaSeries
			if err := manga.LoadFromJSON(metadataPath); err != nil {
				// Log the error but continue with other manga
				continue
			}
			mangas = append(mangas, manga)
		} else {
			// Try to create metadata from directory structure
			if manga, err := mm.CreateMangaFromDirectory(mangaPath); err == nil {
				mangas = append(mangas, manga)
			}
		}
	}

	return mangas, nil
}

// GetMangaByID returns a specific manga by its ID
func (mm *MetadataManager) GetMangaByID(id string) (*MangaSeries, error) {
	// First, try the direct path approach if ID matches directory name
	mangaPath := filepath.Join(mm.RootDir, id)
	metadataPath := filepath.Join(mangaPath, MetadataFileName)

	if _, err := os.Stat(metadataPath); err == nil {
		var manga MangaSeries
		if err := manga.LoadFromJSON(metadataPath); err != nil {
			return nil, err
		}
		return &manga, nil
	}

	// If not found by direct path, scan all manga
	mangas, err := mm.ScanForManga()
	if err != nil {
		return nil, err
	}

	// Look for matching ID
	for i, manga := range mangas {
		if manga.ID == id {
			return &mangas[i], nil
		}
	}

	return nil, NewMangaNotFoundError("no manga with ID: " + id)
}

// CreateMangaFromDirectory attempts to create manga metadata from directory structure
func (mm *MetadataManager) CreateMangaFromDirectory(dirPath string) (MangaSeries, error) {
	manga := MangaSeries{
		ID:          filepath.Base(dirPath),
		Title:       strings.ReplaceAll(filepath.Base(dirPath), "-", " "),
		Description: "No description available",
		Path:        dirPath,
		LastUpdated: time.Now(),
		Status:      "Unknown",
	}

	// Look for a cover image
	files, _ := os.ReadDir(dirPath)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		lower := strings.ToLower(file.Name())
		if strings.Contains(lower, "cover") || lower == "thumbnail.jpg" || lower == "thumbnail.png" {
			manga.CoverImage = file.Name()
			break
		}
	}

	if manga.CoverImage == "" {
		// Use the first image file found as cover
		for _, file := range files {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".jpg" || ext == ".png" || ext == ".jpeg" {
				manga.CoverImage = file.Name()
				break
			}
		}
	}

	// Count chapters
	chapters, _ := mm.ScanForChapters(&manga)
	manga.ChapterCount = len(chapters)

	return manga, nil
}

// ScanForChapters scans a manga directory for chapters
func (mm *MetadataManager) ScanForChapters(manga *MangaSeries) ([]Chapter, error) {
	var chapters []Chapter

	// Read the manga directory
	entries, err := os.ReadDir(manga.Path)
	if err != nil {
		return nil, NewMetadataError("failed to read manga directory: " + err.Error())
	}

	for _, entry := range entries {
		// Skip non-directories and hidden directories
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		chapterPath := filepath.Join(manga.Path, entry.Name())
		metadataPath := filepath.Join(chapterPath, MetadataFileName)

		// If metadata exists, load it
		if _, err := os.Stat(metadataPath); err == nil {
			var chapter Chapter
			if err := chapter.LoadFromJSON(metadataPath); err != nil {
				// Log error but continue
				continue
			}
			chapters = append(chapters, chapter)
		} else {
			// Try to create chapter metadata from directory name
			// Common patterns: chapter-1, ch1, 01, etc.
			if chapter, err := mm.CreateChapterFromDirectory(manga.ID, chapterPath); err == nil {
				chapters = append(chapters, chapter)
			}
		}
	}

	return chapters, nil
}

// CreateChapterFromDirectory attempts to create chapter metadata from directory structure
func (mm *MetadataManager) CreateChapterFromDirectory(mangaID, dirPath string) (Chapter, error) {
	dirName := filepath.Base(dirPath)

	// Try to extract chapter number from directory name
	var chapterNum float64 = 0

	// Remove prefix like "chapter-" or "ch"
	processedName := strings.ToLower(dirName)
	processedName = strings.ReplaceAll(processedName, "chapter-", "")
	processedName = strings.ReplaceAll(processedName, "chapter", "")
	processedName = strings.ReplaceAll(processedName, "ch", "")

	// Try to parse the remaining string as a number
	_, err := json.Marshal(processedName)
	if err == nil {
		// It's a valid JSON string, try to convert to number
		if num, err := jsonNumberToFloat(processedName); err == nil {
			chapterNum = num
		}
	}

	// If we couldn't extract a number, use a default
	if chapterNum == 0 {
		// Just use directory name as ID
		chapterNum = 1
	}

	// Count pages
	var pageCount int
	entries, _ := os.ReadDir(dirPath)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".jpg" || ext == ".png" || ext == ".jpeg" {
			pageCount++
		}
	}

	chapter := Chapter{
		ID:          dirName,
		MangaID:     mangaID,
		Number:      chapterNum,
		Title:       strings.ReplaceAll(dirName, "-", " "),
		ReleaseDate: time.Now(),
		PageCount:   pageCount,
		Path:        dirPath,
	}

	return chapter, nil
}

// Helper function to convert a JSON-encoded string to a float
func jsonNumberToFloat(s string) (float64, error) {
	var num float64
	err := json.Unmarshal([]byte(s), &num)
	return num, err
}
