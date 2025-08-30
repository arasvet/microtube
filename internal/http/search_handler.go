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

type SearchHandler struct {
	UC usecase.SearchUCInterface
}

func (h *SearchHandler) Register(r chi.Router) {
	r.Get("/search", h.searchVideos)
}

// searchVideos обрабатывает GET запрос для поиска видео
func (h *SearchHandler) searchVideos(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Парсим limit и offset с значениями по умолчанию
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Создаем параметры поиска
	params := domain.SearchParams{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	// Выполняем поиск
	results, err := h.UC.SearchVideos(r.Context(), params)
	if err != nil {
		log.Printf("ошибка поиска: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"query":   query,
		"limit":   limit,
		"offset":  offset,
		"total":   len(results),
		"results": results,
	}

	// Отправляем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
