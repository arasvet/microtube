package repo

import (
	"context"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/jackc/pgx/v5"
)

var (
	ErrDuplicate = pgx.ErrNoRows // будем возвращать nil при дубле, поэтому не используется наружу
)

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Store interface {
	Begin(ctx context.Context) (Tx, error)

	// Users
	CreateUser(ctx context.Context, id, email, passHash string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (id string, passHash string, err error)

	// Events
	InsertEvent(ctx context.Context, tx Tx, e domain.Event) (bool, error)
	ExistsEvent(ctx context.Context, event domain.Event) (bool, error)
	UpsertVideoCounters(ctx context.Context, tx Tx, e domain.Event) error
	UpsertVideoDaily(ctx context.Context, tx Tx, e domain.Event) error
	UpdateUserSignalsBestEffort(ctx context.Context, tx Tx, e domain.Event) error

	// Search
	SearchVideos(ctx context.Context, params domain.SearchParams) ([]domain.SearchResult, error)

	// Feeds
	GetPopularVideos(ctx context.Context, limit int) ([]domain.Video, error)
	GetCommentedVideos(ctx context.Context, limit int) ([]domain.Video, error)
	GetRandomVideos(ctx context.Context, limit int) ([]domain.Video, error)

	// Рекомендации
	GetUserTopTags(ctx context.Context, userID string) ([]string, error)
	GetSessionTopTags(ctx context.Context, sessionID string) ([]string, error)
	GetVideosByTags(ctx context.Context, tags []string, limit int) ([]domain.Video, error)
	GetSimilarVideos(ctx context.Context, videoID string, limit int) ([]domain.Video, error)
	GetDiversifiedVideos(ctx context.Context, excludeIDs []string, limit int) ([]domain.Video, error)

	// Статистика
	StatsTotals(ctx context.Context, from, to string) (domain.StatsTotals, error)
	StatsTopVideos(ctx context.Context, from, to string, top int) ([]domain.VideoWithStats, error)
}
