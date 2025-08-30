package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arasvet/microtube/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type StatsHandler struct {
	UC usecase.StatsUCInterface
}

func (h *StatsHandler) Register(r chi.Router) {
	r.Get("/stats/overview", h.overview)
}

func (h *StatsHandler) overview(w http.ResponseWriter, r *http.Request) {
	// Простейшая авторизация: если токен валиден — ок; без токена тоже можно (если нужно — включим требование)
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	topStr := q.Get("top")
	if topStr == "" {
		topStr = "10"
	}
	top, err := strconv.Atoi(topStr)
	if err != nil || top <= 0 {
		top = 10
	}

	var from, to time.Time
	if fromStr == "" || toStr == "" {
		// По умолчанию последние 30 дней
		to = time.Now().UTC().Truncate(24 * time.Hour)
		from = to.AddDate(0, 0, -30)
	} else {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, "invalid from", http.StatusBadRequest)
			return
		}
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, "invalid to", http.StatusBadRequest)
			return
		}
	}

	res, err := h.UC.Overview(r.Context(), from, to, top)
	if err != nil {
		log.Printf("stats overview error: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
