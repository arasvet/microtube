package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Video представляет видео в системе
type Video struct {
	ID          uuid.UUID
	Title       string
	Description string
	Lang        string
	Tags        []string
	DurationS   int
	UploadedAt  time.Time
	AuthorID    *uuid.UUID
}

// SearchResult представляет результат поиска с релевантностью
type SearchResult struct {
	Video Video
	Score float64
}

// SearchParams параметры поиска
type SearchParams struct {
	Query  string
	Limit  int
	Offset int
}

// Validate проверяет корректность параметров поиска
func (sp *SearchParams) Validate() error {
	if sp.Query == "" {
		return ErrInvalidSearchParams
	}
	if sp.Limit <= 0 {
		sp.Limit = 20 // значение по умолчанию
	}
	if sp.Offset < 0 {
		sp.Offset = 0
	}
	if sp.Limit > 100 { // ограничиваем максимальный размер выборки
		sp.Limit = 100
	}
	return nil
}

var ErrInvalidSearchParams = errors.New("invalid search parameters")
