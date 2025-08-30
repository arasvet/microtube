package domain

// FeedType тип фида видео
type FeedType string

const (
	FeedTypePopular   FeedType = "popular"   // популярное видео на основе просмотров и лайков с затуханием по времени
	FeedTypeCommented FeedType = "commented" // прокси по лайкам и завершениям
	FeedTypeRandom    FeedType = "random"    // случайная выборка
)

// FeedParams параметры для получения фида
type FeedParams struct {
	Type  FeedType
	Limit int
}

// Validate проверяет корректность параметров фида
func (fp *FeedParams) Validate() error {
	// Если тип не указан или неверный, используем popular по умолчанию
	switch fp.Type {
	case FeedTypePopular, FeedTypeCommented, FeedTypeRandom:
		// тип корректен
	default:
		fp.Type = FeedTypePopular // используем popular по умолчанию
	}

	if fp.Limit <= 0 {
		fp.Limit = 20 // значение по умолчанию
	}
	if fp.Limit > 100 { // ограничиваем максимальный размер выборки
		fp.Limit = 100
	}
	return nil
}
