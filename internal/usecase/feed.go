package usecase

import (
	"context"
	"fmt"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/repo"
)

// FeedUCInterface интерфейс для тестирования
type FeedUCInterface interface {
	GetFeed(ctx context.Context, params domain.FeedParams) ([]domain.Video, error)
}

type FeedUC struct {
	store repo.Store
}

func NewFeedUC(store repo.Store) *FeedUC {
	return &FeedUC{store: store}
}

// GetFeed возвращает фид видео в зависимости от типа
func (uc *FeedUC) GetFeed(ctx context.Context, params domain.FeedParams) ([]domain.Video, error) {
	// Валидируем параметры фида
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Получаем видео в зависимости от типа фида
	switch params.Type {
	case domain.FeedTypePopular:
		return uc.store.GetPopularVideos(ctx, params.Limit)
	case domain.FeedTypeCommented:
		return uc.store.GetCommentedVideos(ctx, params.Limit)
	case domain.FeedTypeRandom:
		return uc.store.GetRandomVideos(ctx, params.Limit)
	default:
		return nil, fmt.Errorf("unsupported feed type: %s", params.Type)
	}
}
