package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arasvet/microtube/internal/config"
	apihttp "github.com/arasvet/microtube/internal/http"
	"github.com/arasvet/microtube/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.MustLoad()

	// Postgres
	// todo cfg
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, cfg.PostgresURL())
	if err != nil {
		slog.Error("cannot create postgres pool", slog.String("err", err.Error()))
		os.Exit(1)
	}

	if err = dbpool.Ping(ctx); err != nil {
		slog.Error("cannot connect postgres", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer dbpool.Close()

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		DB:   cfg.RedisDB,
	})

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		slog.Error("cannot connect redis", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = rdb.Close()
	}()

	// Repos
	repos := repo.New(dbpool, rdb)

	// Router
	r := chi.NewRouter()
	apihttp.SetupMiddleware(r, cfg.JWTSecret)
	apihttp.SetupRoutes(r, repos, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.APIHttpPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start
	go func() {
		slog.Info("listening", slog.String("addr", ":"+cfg.APIHttpPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	slog.Info("shutting down...")

	// todo cfg
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(ctxShutdown)
}
