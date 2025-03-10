package models

import "fmt"

// Custom error types to provide clearer context for errors

// MangaNotFoundError indicates that a manga series was not found
type MangaNotFoundError struct {
	Message string
}

func (e MangaNotFoundError) Error() string {
	return fmt.Sprintf("manga not found: %s", e.Message)
}

// NewMangaNotFoundError creates a new MangaNotFoundError
func NewMangaNotFoundError(message string) error {
	return MangaNotFoundError{Message: message}
}

// IsMangaNotFoundError checks if an error is a MangaNotFoundError
func IsMangaNotFoundError(err error) bool {
	_, ok := err.(MangaNotFoundError)
	return ok
}

// ChapterNotFoundError indicates that a chapter was not found
type ChapterNotFoundError struct {
	Message string
}

func (e ChapterNotFoundError) Error() string {
	return fmt.Sprintf("chapter not found: %s", e.Message)
}

// NewChapterNotFoundError creates a new ChapterNotFoundError
func NewChapterNotFoundError(message string) error {
	return ChapterNotFoundError{Message: message}
}

// IsChapterNotFoundError checks if an error is a ChapterNotFoundError
func IsChapterNotFoundError(err error) bool {
	_, ok := err.(ChapterNotFoundError)
	return ok
}

// PageNotFoundError indicates that a page was not found
type PageNotFoundError struct {
	Message string
}

func (e PageNotFoundError) Error() string {
	return fmt.Sprintf("page not found: %s", e.Message)
}

// NewPageNotFoundError creates a new PageNotFoundError
func NewPageNotFoundError(message string) error {
	return PageNotFoundError{Message: message}
}

// IsPageNotFoundError checks if an error is a PageNotFoundError
func IsPageNotFoundError(err error) bool {
	_, ok := err.(PageNotFoundError)
	return ok
}

// MetadataError indicates an error with reading or writing metadata
type MetadataError struct {
	Message string
}

func (e MetadataError) Error() string {
	return fmt.Sprintf("metadata error: %s", e.Message)
}

// NewMetadataError creates a new MetadataError
func NewMetadataError(message string) error {
	return MetadataError{Message: message}
}

// IsMetadataError checks if an error is a MetadataError
func IsMetadataError(err error) bool {
	_, ok := err.(MetadataError)
	return ok
}

// ValidationError indicates that a model failed validation
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string) error {
	return ValidationError{Message: message}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}
