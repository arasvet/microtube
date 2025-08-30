package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	defaultVideos = 1200
	defaultEvents = 50000

	tags = []string{
		"go", "microservices", "postgres", "docker", "redis", "kafka",
		"concurrency", "ai", "ml", "cloud", "devops", "frontend",
		"backend", "testing", "lms",
	}

	eventTypes = []string{"view_start", "view_complete", "like", "search_query", "click_result"}
)

func main() {
	ctx := context.Background()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	rand.Seed(time.Now().UnixNano())

	// количество видео/событий берём из ENV или дефолтные
	videoCount := getEnvInt("VIDEOS", defaultVideos)
	eventCount := getEnvInt("EVENTS", defaultEvents)

	fmt.Printf("==> seeding %d videos and %d events\n", videoCount, eventCount)

	// сидим демо-пользователей
	fmt.Println("==> seeding users")
	for i := 0; i < 10; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		_, err := pool.Exec(ctx, `
			INSERT INTO app.users (id, email, pass_hash, created_at)
			VALUES ($1, $2, $3, now())
			ON CONFLICT (email) DO NOTHING`,
			uuid.New(), email, "hashed_password")
		if err != nil {
			panic(err)
		}
	}

	// сидим видео
	fmt.Println("==> seeding videos")
	for i := 0; i < videoCount; i++ {
		id := uuid.New()
		title := fmt.Sprintf("Video %d about %s", i, tags[rand.Intn(len(tags))])
		desc := fmt.Sprintf("Description for video %d", i)
		duration := rand.Intn(600) + 60 // 1–10 мин
		tagset := fmt.Sprintf("{%s,%s}", tags[rand.Intn(len(tags))], tags[rand.Intn(len(tags))])

		_, err := pool.Exec(ctx, `
			INSERT INTO app.videos (id, title, description, lang, tags, duration_s, uploaded_at)
			VALUES ($1, $2, $3, 'en', $4::text[], $5, now() - ($6 * interval '1 day'))
			ON CONFLICT (id) DO NOTHING`,
			id, title, desc, tagset, duration, rand.Intn(365))
		if err != nil {
			panic(err)
		}
		if i > 0 && i%200 == 0 {
			fmt.Printf("inserted %d videos\n", i)
		}
	}

	// сидим события
	fmt.Println("==> seeding events")
	for i := 0; i < eventCount; i++ {
		eventID := uuid.New()
		sessionID := fmt.Sprintf("sess-%d", rand.Intn(1000))
		etype := eventTypes[rand.Intn(len(eventTypes))]

		_, err := pool.Exec(ctx, `
			INSERT INTO app.events (event_id, ts, type, session_id, video_id, query, dwell_ms)
			VALUES ($1, now() - ($2 * interval '1 minute'), $3, $4, $5, $6, $7)
			ON CONFLICT (event_id) DO NOTHING`,
			eventID,
			rand.Intn(60*24*30), // минуты за последний месяц
			etype,
			sessionID,
			randomVideoID(pool, ctx, videoCount),
			randomQuery(etype),
			rand.Intn(10000),
		)
		if err != nil {
			panic(err)
		}
		if i > 0 && i%10000 == 0 {
			fmt.Printf("inserted %d events\n", i)
		}
	}

	fmt.Println("==> seeding done")
}

func randomQuery(eventType string) *string {
	if eventType == "search_query" {
		q := []string{"golang", "docker", "postgres", "redis", "ai", "lms"}[rand.Intn(6)]
		return &q
	}
	return nil
}

func randomVideoID(pool *pgxpool.Pool, ctx context.Context, max int) uuid.UUID {
	var id uuid.UUID
	err := pool.QueryRow(ctx,
		"SELECT id FROM app.videos OFFSET floor(random()*$1) LIMIT 1", max).Scan(&id)
	if err != nil {
		return uuid.New()
	}
	return id
}

func getEnvInt(key string, def int) int {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}
