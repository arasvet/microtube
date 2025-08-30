package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/arasvet/microtube/internal/repo"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	UC *usecase.AuthUC
}

func (h *AuthHandler) Register(r chi.Router) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
}

type registerIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginIn registerIn

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var in registerIn

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

	id, err := h.UC.Register(r.Context(), in.Email, in.Password)
	if err != nil {
		// если регистрация не вернула id (email существует) — 409
		if errors.Is(err, repo.ErrDuplicate) || id == "" {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{"user_id": id})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var in loginIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

	tok, err := h.UC.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		log.Printf("login failed: %v", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": tok})
}
