package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/repo"
)

// RecommendationsUCInterface интерфейс для тестирования
type RecommendationsUCInterface interface {
	GetRecommendations(ctx context.Context, params domain.RecommendationParams) ([]domain.RecommendationResult, error)
}

type RecommendationsUC struct {
	store repo.Store
}

func NewRecommendationsUC(store repo.Store) *RecommendationsUC {
	return &RecommendationsUC{store: store}
}

// GetRecommendations возвращает персональные или холодные рекомендации
func (uc *RecommendationsUC) GetRecommendations(ctx context.Context, params domain.RecommendationParams) ([]domain.RecommendationResult, error) {
	// Валидируем параметры
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Определяем тип рекомендаций
	var recType domain.RecommendationType
	if params.UserID != nil {
		recType = domain.RecommendationTypePersonal
	} else {
		recType = domain.RecommendationTypeCold
	}

	// Получаем рекомендации в зависимости от типа
	var results []domain.RecommendationResult
	var err error

	switch recType {
	case domain.RecommendationTypePersonal:
		results, err = uc.getPersonalRecommendations(ctx, *params.UserID, params.Limit)
	case domain.RecommendationTypeCold:
		results, err = uc.getColdRecommendations(ctx, *params.SessionID, params.Limit)
	default:
		results, err = uc.getMixedRecommendations(ctx, params, params.Limit)
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}

// getPersonalRecommendations возвращает персональные рекомендации для авторизованного пользователя
func (uc *RecommendationsUC) getPersonalRecommendations(ctx context.Context, userID string, limit int) ([]domain.RecommendationResult, error) {
	var results []domain.RecommendationResult

	// 1. Получаем топ теги пользователя (40% рекомендаций)
	userTags, err := uc.store.GetUserTopTags(ctx, userID)
	if err == nil && len(userTags) > 0 {
		tagLimit := limit * 4 / 10 // 40%
		videos, err := uc.store.GetVideosByTags(ctx, userTags, tagLimit)
		if err == nil {
			for _, video := range videos {
				results = append(results, domain.RecommendationResult{
					Video:  video,
					Reason: domain.ReasonUserTags,
					Score:  0.9, // высокий score для персональных рекомендаций
				})
			}
		}
	}

	// 2. Добавляем популярные видео (30% рекомендаций)
	popularLimit := limit * 3 / 10 // 30%
	if len(results) < limit {
		popularVideos, err := uc.store.GetPopularVideos(ctx, popularLimit)
		if err == nil {
			for _, video := range popularVideos {
				// Проверяем, что видео еще не добавлено
				if !uc.videoExists(results, video.ID.String()) {
					results = append(results, domain.RecommendationResult{
						Video:  video,
						Reason: domain.ReasonPopular,
						Score:  0.7,
					})
				}
			}
		}
	}

	// 3. Добавляем диверсификацию (20% рекомендаций)
	diversifyLimit := limit * 2 / 10 // 20%
	if len(results) < limit {
		excludeIDs := uc.getVideoIDs(results)
		diversifiedVideos, err := uc.store.GetDiversifiedVideos(ctx, excludeIDs, diversifyLimit)
		if err == nil {
			for _, video := range diversifiedVideos {
				if !uc.videoExists(results, video.ID.String()) {
					results = append(results, domain.RecommendationResult{
						Video:  video,
						Reason: domain.ReasonDiversify,
						Score:  0.6,
					})
				}
			}
		}
	}

	// 4. Добавляем exploration (10% рекомендаций)
	explorationLimit := limit * 1 / 10 // 10%
	if len(results) < limit {
		excludeIDs := uc.getVideoIDs(results)
		explorationVideos, err := uc.store.GetDiversifiedVideos(ctx, excludeIDs, explorationLimit)
		if err == nil {
			for _, video := range explorationVideos {
				if !uc.videoExists(results, video.ID.String()) {
					results = append(results, domain.RecommendationResult{
						Video:  video,
						Reason: domain.ReasonExploration,
						Score:  0.5,
					})
				}
			}
		}
	}

	// Ограничиваем результат
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// getColdRecommendations возвращает холодные рекомендации для гостей
func (uc *RecommendationsUC) getColdRecommendations(ctx context.Context, sessionID string, limit int) ([]domain.RecommendationResult, error) {
	var results []domain.RecommendationResult

	// 1. Популярные видео (50% рекомендаций)
	popularLimit := limit * 5 / 10
	popularVideos, err := uc.store.GetPopularVideos(ctx, popularLimit)
	if err == nil {
		for _, video := range popularVideos {
			results = append(results, domain.RecommendationResult{
				Video:  video,
				Reason: domain.ReasonPopular,
				Score:  0.8,
			})
		}
	}

	// 2. Попробуем получить теги сессии (30% рекомендаций)
	if len(results) < limit {
		sessionTags, err := uc.store.GetSessionTopTags(ctx, sessionID)
		if err == nil && len(sessionTags) > 0 {
			tagLimit := limit * 3 / 10
			videos, err := uc.store.GetVideosByTags(ctx, sessionTags, tagLimit)
			if err == nil {
				for _, video := range videos {
					if !uc.videoExists(results, video.ID.String()) {
						results = append(results, domain.RecommendationResult{
							Video:  video,
							Reason: domain.ReasonUserTags,
							Score:  0.7,
						})
					}
				}
			}
		}
	}

	// 3. Диверсификация (20% рекомендаций)
	if len(results) < limit {
		diversifyLimit := limit - len(results)
		excludeIDs := uc.getVideoIDs(results)
		diversifiedVideos, err := uc.store.GetDiversifiedVideos(ctx, excludeIDs, diversifyLimit)
		if err == nil {
			for _, video := range diversifiedVideos {
				if !uc.videoExists(results, video.ID.String()) {
					results = append(results, domain.RecommendationResult{
						Video:  video,
						Reason: domain.ReasonDiversify,
						Score:  0.6,
					})
				}
			}
		}
	}

	// Ограничиваем результат
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// getMixedRecommendations возвращает смешанные рекомендации
func (uc *RecommendationsUC) getMixedRecommendations(ctx context.Context, params domain.RecommendationParams, limit int) ([]domain.RecommendationResult, error) {
	// Для смешанных рекомендаций используем комбинацию подходов
	rand.Seed(time.Now().UnixNano())

	if rand.Float32() < 0.5 {
		// 50% вероятность персональных рекомендаций
		if params.UserID != nil {
			return uc.getPersonalRecommendations(ctx, *params.UserID, limit)
		}
	}

	// Fallback к холодным рекомендациям
	if params.SessionID != nil {
		return uc.getColdRecommendations(ctx, *params.SessionID, limit)
	}

	// Если ничего не подходит, возвращаем популярные
	popularVideos, err := uc.store.GetPopularVideos(ctx, limit)
	if err != nil {
		return nil, err
	}

	var results []domain.RecommendationResult
	for _, video := range popularVideos {
		results = append(results, domain.RecommendationResult{
			Video:  video,
			Reason: domain.ReasonPopular,
			Score:  0.7,
		})
	}

	return results, nil
}

// Вспомогательные методы
func (uc *RecommendationsUC) videoExists(results []domain.RecommendationResult, videoID string) bool {
	for _, result := range results {
		if result.Video.ID.String() == videoID {
			return true
		}
	}
	return false
}

func (uc *RecommendationsUC) getVideoIDs(results []domain.RecommendationResult) []string {
	ids := make([]string, 0, len(results))
	for _, result := range results {
		ids = append(ids, result.Video.ID.String())
	}
	return ids
}
