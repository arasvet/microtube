package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EventsHandler struct {
	UC *usecase.EventsUC
}

func (h *EventsHandler) Register(r chi.Router) {
	r.Post("/events", h.postEvent)
}

type postEventIn struct {
	EventID   string    `json:"event_id"`
	TS        time.Time `json:"ts"`
	Type      string    `json:"type"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	VideoID   string    `json:"video_id"`
	Query     string    `json:"query"`
	DwellMs   int       `json:"dwell_ms"`
}

func (h *EventsHandler) postEvent(w http.ResponseWriter, r *http.Request) {
	var in postEventIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad JSON body", http.StatusBadRequest)
		return
	}

	evID, err := uuid.Parse(in.EventID)
	if err != nil || in.SessionID == "" || in.Type == "" || in.TS.IsZero() {
		http.Error(w, "invalid fields", http.StatusUnprocessableEntity)
		return
	}

	var uid uuid.UUID
	if in.UserID != "" {
		u, err := uuid.Parse(in.UserID)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusUnprocessableEntity)
			return
		}
		uid = u
	}
	var vid uuid.UUID
	if in.VideoID != "" {
		v, err := uuid.Parse(in.VideoID)
		if err != nil {
			http.Error(w, "invalid video_id", http.StatusUnprocessableEntity)
			return
		}
		vid = v
	}

	e := domain.Event{
		EventID:   evID,
		TS:        in.TS,
		Type:      domain.EventType(in.Type),
		SessionID: in.SessionID,
		UserID:    uid,
		VideoID:   vid,
		Query:     in.Query,
		DwellMs:   in.DwellMs,
	}
	if err := e.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ingest, err := h.UC.Ingest(r.Context(), e)
	if err != nil {
		log.Printf("failed to ingest event: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	// 201 для нового события, 200 для дубля — здесь не различаем, это ок.
	if !ingest.Inserted {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
