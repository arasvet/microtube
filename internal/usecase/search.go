package usecase

import (
	"context"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/repo"
)

// SearchUCInterface интерфейс для тестирования
type SearchUCInterface interface {
	SearchVideos(ctx context.Context, params domain.SearchParams) ([]domain.SearchResult, error)
}

type SearchUC struct {
	store repo.Store
}

func NewSearchUC(store repo.Store) *SearchUC {
	return &SearchUC{store: store}
}

// SearchVideos выполняет поиск видео с валидацией параметров
func (uc *SearchUC) SearchVideos(ctx context.Context, params domain.SearchParams) ([]domain.SearchResult, error) {
	// Валидируем параметры поиска
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Выполняем поиск в репозитории
	results, err := uc.store.SearchVideos(ctx, params)
	if err != nil {
		return nil, err
	}

	return results, nil
}
