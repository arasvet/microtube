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

type RecommendationsHandler struct {
	UC usecase.RecommendationsUCInterface
}

func (h *RecommendationsHandler) Register(r chi.Router) {
	r.Get("/recommendations", h.getRecommendations)
}

// getRecommendations обрабатывает GET запрос для получения рекомендаций
func (h *RecommendationsHandler) getRecommendations(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	userID := r.URL.Query().Get("user_id")
	sessionID := r.URL.Query().Get("session_id")

	// Парсим limit с значением по умолчанию
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Создаем параметры рекомендаций
	params := domain.RecommendationParams{
		Limit: limit,
	}

	// Устанавливаем user_id или session_id
	if userID != "" {
		params.UserID = &userID
	} else if sessionID != "" {
		params.SessionID = &sessionID
	}

	// Получаем рекомендации
	results, err := h.UC.GetRecommendations(r.Context(), params)
	if err != nil {
		log.Printf("ошибка получения рекомендаций: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Определяем тип рекомендаций для ответа
	var recType string
	if params.UserID != nil {
		recType = "personal"
	} else if params.SessionID != nil {
		recType = "cold"
	} else {
		recType = "mixed"
	}

	// Формируем ответ
	response := map[string]interface{}{
		"type":            recType,
		"limit":           limit,
		"total":           len(results),
		"recommendations": results,
	}

	// Отправляем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}
