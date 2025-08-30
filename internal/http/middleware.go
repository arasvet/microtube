package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ctxKey string

const userIDCtxKey ctxKey = "user_id"

func SetupMiddleware(r chi.Router, jwtSecret string) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(JWTAuthMiddleware(jwtSecret))
}

// JWTAuthMiddleware проверяет заголовок Authorization: Bearer <jwt>
// Валидирует подпись HS256 и кладёт subject (sub) в контекст как user_id.
func JWTAuthMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				// без токена считаем гостем
				next.ServeHTTP(w, r)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			sub, err := validateAndExtractSubHS256(token, secret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDCtxKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateAndExtractSubHS256 валидирует JWT (header.payload.signature) HS256 и возвращает sub.
func validateAndExtractSubHS256(token, secret string) (string, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return "", errors.New("bad token segments")
	}
	headEnc, payloadEnc, sigEnc := segments[0], segments[1], segments[2]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(headEnc + "." + payloadEnc))
	expectedSig := mac.Sum(nil)

	sig, err := base64.RawURLEncoding.DecodeString(sigEnc)
	if err != nil {
		return "", err
	}
	if !hmac.Equal(sig, expectedSig) {
		return "", errors.New("signature mismatch")
	}

	// decode header to ensure HS256
	headBytes, err := base64.RawURLEncoding.DecodeString(headEnc)
	if err != nil {
		return "", err
	}
	var head map[string]any
	if err := json.Unmarshal(headBytes, &head); err != nil {
		return "", err
	}
	if alg, _ := head["alg"].(string); !strings.EqualFold(alg, "HS256") {
		return "", errors.New("unsupported alg")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadEnc)
	if err != nil {
		return "", err
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", err
	}
	// Проверяем exp, если есть
	if expVal, ok := payload["exp"]; ok {
		switch v := expVal.(type) {
		case float64:
			// UNIX seconds
			if int64(v) <= time.Now().Unix() {
				return "", errors.New("token expired")
			}
		}
	}
	sub, _ := payload["sub"].(string)
	if sub == "" {
		return "", errors.New("empty sub")
	}
	return sub, nil
}

// UserIDFromContext возвращает user_id (sub) из контекста, если авторизован
func UserIDFromContext(r *http.Request) (string, bool) {
	v := r.Context().Value(userIDCtxKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}
