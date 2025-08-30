package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/arasvet/microtube/internal/config"
	"github.com/arasvet/microtube/internal/idem"
	"github.com/arasvet/microtube/internal/repo"
	"github.com/arasvet/microtube/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router, repos *repo.Repositories, cfg config.Config) {
	// Корневая страница - перенаправление на документацию
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, map[string]string{"status": "ok"}) })

	// OpenAPI статика из embed
	r.Get("/openapi.yaml", serveOpenAPI)

	// Swagger UI для просмотра API документации
	r.Get("/docs", serveSwaggerUI)

	// init
	authUC := usecase.NewAuthUC(cfg, repos.Postgres, cfg.JWTSecret)
	eventsUC := usecase.NewEventsUC(repos.Postgres, idem.New(repos.Redis.Client()))
	searchUC := usecase.NewSearchUC(repos.Postgres)
	feedUC := usecase.NewFeedUC(repos.Postgres)
	recommendationsUC := usecase.NewRecommendationsUC(repos.Postgres)
	statsUC := usecase.NewStatsUC(repos.Postgres)

	// register routes
	(&AuthHandler{UC: authUC}).Register(r)
	(&EventsHandler{UC: eventsUC}).Register(r)
	(&SearchHandler{UC: searchUC}).Register(r)
	(&FeedHandler{UC: feedUC}).Register(r)
	(&RecommendationsHandler{UC: recommendationsUC}).Register(r)

	// auth
	r.Group(func(ar chi.Router) {
		ar.Use(adminOnlyMiddleware())
		(&StatsHandler{UC: statsUC}).Register(ar)
	})
}

// todo
func adminOnlyMiddleware() func(next http.Handler) http.Handler {
	admins := strings.Split(os.Getenv("ADMINS"), ",")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r)
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			for _, a := range admins {
				if strings.TrimSpace(a) != "" && strings.TrimSpace(a) == userID {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
