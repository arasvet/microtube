package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type FeedHandler struct {
	UC usecase.FeedUCInterface
}

func (h *FeedHandler) Register(r chi.Router) {
	r.Get("/videos/feed", h.getFeed)
}

// getFeed обрабатывает GET запрос для получения фида видео
func (h *FeedHandler) getFeed(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	feedType := r.URL.Query().Get("type")
	if feedType == "" {
		feedType = string(domain.FeedTypePopular) // значение по умолчанию
	}

	// Парсим limit с значением по умолчанию
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Создаем параметры фида
	params := domain.FeedParams{
		Type:  domain.FeedType(feedType),
		Limit: limit,
	}

	// Валидируем параметры (тип будет исправлен на popular если неверный)
	if err := params.Validate(); err != nil {
		log.Printf("ошибка валидации параметров фида: %v", err)
		http.Error(w, "неверные параметры", http.StatusBadRequest)
		return
	}

	// Получаем фид
	videos, err := h.UC.GetFeed(r.Context(), params)
	if err != nil {
		log.Printf("ошибка получения фида: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Формируем ответ (используем валидированный тип)
	response := map[string]interface{}{
		"type":   string(params.Type),
		"limit":  params.Limit,
		"total":  len(videos),
		"videos": videos,
	}

	// Отправляем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}
