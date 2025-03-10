package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// MetadataFileName is the standard name for metadata files
	MetadataFileName = "metadata.json"
)

// We'll use a package-level logger for convenience
var logger *zap.Logger

// init sets up zap logger. Replace zap.NewDevelopment() with zap.NewProduction() if desired.
func init() {
	logr, _ := zap.NewDevelopment()
	logger = logr
}

// MetadataManager provides utilities for managing metadata
type MetadataManager struct {
	RootDir string // Root directory for manga storage
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(rootDir string) *MetadataManager {
	logger.Info("NewMetadataManager called",
		zap.String("RootDir", rootDir),
	)
	return &MetadataManager{
		RootDir: rootDir,
	}
}

// ScanForManga scans the root directory for manga series
func (mm *MetadataManager) ScanForManga() ([]MangaSeries, error) {
	logger.Info("ScanForManga called",
		zap.String("RootDir", mm.RootDir),
	)

	var mangas []MangaSeries

	// Read the root directory
	dirs, err := os.ReadDir(mm.RootDir)
	if err != nil {
		logger.Error("Failed to read root directory",
			zap.Error(err),
		)
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
			logger.Info("Found metadata file",
				zap.String("mangaPath", mangaPath),
				zap.String("metadataPath", metadataPath),
			)

			var manga MangaSeries
			if err := manga.LoadFromJSON(metadataPath); err != nil {
				// Log the error but continue with other manga
				logger.Warn("Failed to load metadata",
					zap.String("metadataPath", metadataPath),
					zap.Error(err),
				)
				continue
			}
			mangas = append(mangas, manga)
		} else {
			// Try to create metadata from directory structure
			logger.Info("No metadata file found; creating from directory",
				zap.String("mangaPath", mangaPath),
			)

			if manga, err := mm.CreateMangaFromDirectory(mangaPath); err == nil {
				mangas = append(mangas, manga)
			} else {
				logger.Warn("Failed to create manga from directory",
					zap.String("mangaPath", mangaPath),
					zap.Error(err),
				)
			}
		}
	}

	logger.Info("ScanForManga complete",
		zap.Int("mangaCount", len(mangas)),
	)
	return mangas, nil
}

// GetMangaByID returns a specific manga by its ID
func (mm *MetadataManager) GetMangaByID(id string) (*MangaSeries, error) {
	logger.Info("GetMangaByID called",
		zap.String("id", id),
	)

	// First, try the direct path approach if ID matches directory name
	mangaPath := filepath.Join(mm.RootDir, id)
	metadataPath := filepath.Join(mangaPath, MetadataFileName)

	if _, err := os.Stat(metadataPath); err == nil {
		logger.Info("Found metadata file for requested ID",
			zap.String("id", id),
			zap.String("metadataPath", metadataPath),
		)

		var manga MangaSeries
		if err := manga.LoadFromJSON(metadataPath); err != nil {
			logger.Error("Failed to load metadata",
				zap.String("id", id),
				zap.String("metadataPath", metadataPath),
				zap.Error(err),
			)
			return nil, err
		}
		return &manga, nil
	}

	// If not found by direct path, scan all manga
	logger.Info("No direct metadata file for ID; scanning all manga",
		zap.String("id", id),
	)

	mangas, err := mm.ScanForManga()
	if err != nil {
		return nil, err
	}

	// Look for matching ID
	for i, manga := range mangas {
		if manga.ID == id {
			logger.Info("Found manga after scanning",
				zap.String("id", id),
			)
			return &mangas[i], nil
		}
	}

	logger.Warn("No manga found with that ID",
		zap.String("id", id),
	)
	return nil, NewMangaNotFoundError("no manga with ID: " + id)
}

// CreateMangaFromDirectory attempts to create manga metadata from directory structure
func (mm *MetadataManager) CreateMangaFromDirectory(dirPath string) (MangaSeries, error) {
	logger.Info("CreateMangaFromDirectory called",
		zap.String("dirPath", dirPath),
	)

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
			logger.Info("Cover image found",
				zap.String("coverImage", file.Name()),
				zap.String("dirPath", dirPath),
			)
			break
		}
	}

	if manga.CoverImage == "" {
		// Use the first image file found as cover
		for _, file := range files {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".jpg" || ext == ".png" || ext == ".jpeg" {
				manga.CoverImage = file.Name()
				logger.Info("Using first image as cover",
					zap.String("coverImage", file.Name()),
					zap.String("dirPath", dirPath),
				)
				break
			}
		}
	}

	// Count chapters
	chapters, _ := mm.ScanForChapters(&manga)
	manga.ChapterCount = len(chapters)
	logger.Info("Created MangaSeries from directory",
		zap.String("mangaID", manga.ID),
		zap.Int("chapterCount", manga.ChapterCount),
	)

	return manga, nil
}

// ScanForChapters scans a manga directory for chapters
func (mm *MetadataManager) ScanForChapters(manga *MangaSeries) ([]Chapter, error) {
	logger.Info("ScanForChapters called",
		zap.String("mangaID", manga.ID),
		zap.String("mangaPath", manga.Path),
	)

	var chapters []Chapter

	// Read the manga directory
	entries, err := os.ReadDir(manga.Path)
	if err != nil {
		logger.Error("Failed to read manga directory",
			zap.String("mangaPath", manga.Path),
			zap.Error(err),
		)
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
			logger.Info("Found chapter metadata",
				zap.String("chapterPath", chapterPath),
				zap.String("metadataPath", metadataPath),
			)

			var chapter Chapter
			if err := chapter.LoadFromJSON(metadataPath); err != nil {
				logger.Warn("Failed to load chapter metadata",
					zap.String("metadataPath", metadataPath),
					zap.Error(err),
				)
				continue
			}
			chapters = append(chapters, chapter)
		} else {
			// Try to create chapter metadata from directory name
			logger.Info("No metadata for chapter, creating from directory",
				zap.String("chapterPath", chapterPath),
			)
			if chapter, err := mm.CreateChapterFromDirectory(manga.ID, chapterPath); err == nil {
				chapters = append(chapters, chapter)
			} else {
				logger.Warn("Failed to create chapter from directory",
					zap.String("chapterPath", chapterPath),
					zap.Error(err),
				)
			}
		}
	}

	logger.Info("ScanForChapters complete",
		zap.String("mangaID", manga.ID),
		zap.Int("chapterCount", len(chapters)),
	)
	return chapters, nil
}

// CreateChapterFromDirectory attempts to create chapter metadata from directory structure
func (mm *MetadataManager) CreateChapterFromDirectory(mangaID, dirPath string) (Chapter, error) {
	dirName := filepath.Base(dirPath)
	logger.Info("CreateChapterFromDirectory called",
		zap.String("mangaID", mangaID),
		zap.String("dirPath", dirPath),
		zap.String("dirName", dirName),
	)

	var chapterNum float64 = 0
	processedName := strings.ToLower(dirName)
	processedName = strings.ReplaceAll(processedName, "chapter-", "")
	processedName = strings.ReplaceAll(processedName, "chapter", "")
	processedName = strings.ReplaceAll(processedName, "ch", "")

	_, err := json.Marshal(processedName)
	if err == nil {
		if num, err := jsonNumberToFloat(processedName); err == nil {
			chapterNum = num
		}
	}

	if chapterNum == 0 {
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

	logger.Info("CreateChapterFromDirectory complete",
		zap.String("chapterID", chapter.ID),
		zap.String("mangaID", chapter.MangaID),
		zap.Float64("chapterNumber", chapter.Number),
		zap.Int("pageCount", pageCount),
	)

	return chapter, nil
}

func jsonNumberToFloat(s string) (float64, error) {
	var num float64
	err := json.Unmarshal([]byte(s), &num)
	return num, err
}
