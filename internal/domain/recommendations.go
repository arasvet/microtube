package domain

import "errors"

// RecommendationType тип рекомендации
type RecommendationType string

const (
	RecommendationTypePersonal RecommendationType = "personal" // персональные рекомендации
	RecommendationTypeCold     RecommendationType = "cold"     // холодные рекомендации для гостей
	RecommendationTypeMixed    RecommendationType = "mixed"    // смешанные рекомендации
)

// RecommendationParams параметры для получения рекомендаций
type RecommendationParams struct {
	UserID    *string // идентификатор пользователя (для авторизованных)
	SessionID *string // идентификатор сессии (для гостей)
	Limit     int     // количество рекомендаций
}

// Validate проверяет корректность параметров рекомендаций
func (rp *RecommendationParams) Validate() error {
	// Должен быть указан либо user_id, либо session_id
	if rp.UserID == nil && rp.SessionID == nil {
		return ErrInvalidRecommendationParams
	}

	if rp.Limit <= 0 {
		rp.Limit = 20 // значение по умолчанию
	}
	if rp.Limit > 100 { // ограничиваем максимальный размер выборки
		rp.Limit = 100
	}
	return nil
}

// RecommendationReason причина рекомендации
type RecommendationReason string

const (
	ReasonPopular     RecommendationReason = "popular"     // популярное видео
	ReasonUserTags    RecommendationReason = "user_tags"   // по тегам пользователя
	ReasonSimilar     RecommendationReason = "similar"     // похожее на просмотренное
	ReasonDiversify   RecommendationReason = "diversify"   // для диверсификации
	ReasonExploration RecommendationReason = "exploration" // случайное исследование
)

// RecommendationResult результат рекомендации с объяснением
type RecommendationResult struct {
	Video  Video
	Reason RecommendationReason
	Score  float64 // релевантность рекомендации
}

var ErrInvalidRecommendationParams = errors.New("user_id or session_id must be specified")
