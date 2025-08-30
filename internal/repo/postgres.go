package repo

import (
	"context"
	"errors"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTx struct{ tx pgx.Tx }

func (t *PostgresTx) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t *PostgresTx) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

type PostgresRepo struct {
	DB *pgxpool.Pool
}

// Begin transaction with reasonable settings
func (r *PostgresRepo) Begin(ctx context.Context) (Tx, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return nil, err
	}
	return &PostgresTx{tx: tx}, nil
}

func (r *PostgresRepo) InsertEvent(ctx context.Context, tx Tx, e domain.Event) (bool, error) {
	cmd, err := tx.(*PostgresTx).tx.Exec(ctx, `
		INSERT INTO app.events(event_id, ts, type, session_id, user_id, video_id, query, dwell_ms)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (event_id) DO NOTHING
	`, e.EventID, e.TS.UTC(), e.Type, e.SessionID, e.UserID, e.VideoID, e.Query, e.DwellMs)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 1, nil
}

func (r *PostgresRepo) ExistsEvent(ctx context.Context, event domain.Event) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE event_id = $1)`, event.EventID.String()).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo) UpsertVideoCounters(ctx context.Context, tx Tx, e domain.Event) error {
	if e.VideoID.String() == "" {
		return nil
	}
	viewsInc, completesInc, likesInc := 0, 0, 0
	switch e.Type {
	case domain.EventViewStart:
		viewsInc = 1
	case domain.EventViewComplete:
		completesInc = 1
	case domain.EventLike:
		likesInc = 1
	default:
		// no-op
	}
	_, err := tx.(*PostgresTx).tx.Exec(ctx, `
		INSERT INTO app.video_counters(video_id, views, completes, likes, last_event_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (video_id) DO UPDATE
		SET views = app.video_counters.views + EXCLUDED.views,
		    completes = app.video_counters.completes + EXCLUDED.completes,
		    likes = app.video_counters.likes + EXCLUDED.likes,
		    last_event_at = GREATEST(app.video_counters.last_event_at, EXCLUDED.last_event_at)
	`, e.VideoID, viewsInc, completesInc, likesInc, e.TS.UTC())
	return err
}

func (r *PostgresRepo) UpsertVideoDaily(ctx context.Context, tx Tx, e domain.Event) error {
	if e.VideoID.String() == "" {
		return nil
	}
	day := e.TS.UTC().Truncate(24 * time.Hour)
	viewsInc, completesInc, likesInc, clicksInc, imprInc, dwellInc := 0, 0, 0, 0, 0, 0
	switch e.Type {
	case domain.EventViewStart:
		viewsInc = 1
	case domain.EventViewComplete:
		completesInc = 1
	case domain.EventLike:
		likesInc = 1
	case domain.EventClickResult:
		clicksInc = 1
	case domain.EventSearchQuery:
		imprInc = 1 // считаем показ выдачи как impression
	}
	if e.DwellMs != 0 {
		dwellInc = e.DwellMs
	}
	_, err := tx.(*PostgresTx).tx.Exec(ctx, `
		INSERT INTO app.video_daily(video_id, day, views, completes, likes, clicks, impressions, dwell_ms_sum)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (video_id, day) DO UPDATE
		SET views = app.video_daily.views + EXCLUDED.views,
		    completes = app.video_daily.completes + EXCLUDED.completes,
		    likes = app.video_daily.likes + EXCLUDED.likes,
		    clicks = app.video_daily.clicks + EXCLUDED.clicks,
		    impressions = app.video_daily.impressions + EXCLUDED.impressions,
		    dwell_ms_sum = app.video_daily.dwell_ms_sum + EXCLUDED.dwell_ms_sum
	`, e.VideoID, day, viewsInc, completesInc, likesInc, clicksInc, imprInc, dwellInc)
	return err
}

func (r *PostgresRepo) UpdateUserSignalsBestEffort(ctx context.Context, tx Tx, e domain.Event) error {
	// для теста: если есть video_id — просто обновим last_seen_at (теги прикрутим позже)
	if e.UserID.String() == "" && e.SessionID == "" {
		return nil
	}
	key := e.SessionID
	if e.UserID.String() != "" {
		key = e.UserID.String()
	}
	_, err := tx.(*PostgresTx).tx.Exec(ctx, `
		INSERT INTO app.user_signals(user_or_session, last_seen_at)
		VALUES ($1, $2)
		ON CONFLICT (user_or_session) DO UPDATE
		SET last_seen_at = GREATEST(app.user_signals.last_seen_at, EXCLUDED.last_seen_at)
	`, key, e.TS.UTC())
	// best-effort: проглатываем serialization ошибки с ретраем наверху (позже)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return nil
}

// SearchVideos выполняет полнотекстовый поиск по видео с поддержкой FTS и trigram
func (r *PostgresRepo) SearchVideos(ctx context.Context, params domain.SearchParams) ([]domain.SearchResult, error) {
	// Комбинированный поиск: FTS + trigram для лучших результатов
	query := `
		WITH search_results AS (
			SELECT 
				v.id,
				v.title,
				v.description,
				v.lang,
				v.tags,
				v.duration_s,
				v.uploaded_at,
				v.author_id,
				-- FTS релевантность
				COALESCE(ts_rank_cd(v.fts_tsv, plainto_tsquery('simple', $1)), 0) as fts_score,
				-- Trigram релевантность для опечаток
				COALESCE(GREATEST(
					similarity(immutable_unaccent(v.title), immutable_unaccent($1)),
					similarity(immutable_unaccent(v.description), immutable_unaccent($1))
				), 0) as trigram_score
			FROM app.videos v
			WHERE 
				-- FTS поиск
				v.fts_tsv @@ plainto_tsquery('simple', $1)
				OR 
				-- Trigram поиск для опечаток (если FTS не дал результатов)
				(immutable_unaccent(v.title) % immutable_unaccent($1) 
				 OR immutable_unaccent(v.description) % immutable_unaccent($1))
		)
		SELECT 
			id, title, description, lang, tags, duration_s, uploaded_at, author_id,
			-- Комбинированный score: FTS имеет больший вес, trigram дополняет
			(fts_score * 0.7 + trigram_score * 0.3) as combined_score
		FROM search_results
		ORDER BY combined_score DESC, uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.Query(ctx, query, params.Query, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.SearchResult
	for rows.Next() {
		var video domain.Video
		var score float64

		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
			&score,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, domain.SearchResult{
			Video: video,
			Score: score,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetPopularVideos возвращает популярные видео на основе просмотров и лайков с затуханием по времени
func (r *PostgresRepo) GetPopularVideos(ctx context.Context, limit int) ([]domain.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
		FROM app.videos v
		LEFT JOIN app.video_counters vc ON v.id = vc.video_id
		ORDER BY 
			-- Популярность с затуханием по времени (более новые видео получают бонус)
			COALESCE(vc.views, 0) * 0.4 + 
			COALESCE(vc.likes, 0) * 0.6 + 
			EXTRACT(EPOCH FROM (now() - v.uploaded_at)) / 86400 * 0.01 DESC
		LIMIT $1
	`

	rows, err := r.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// GetCommentedVideos возвращает видео с прокси по лайкам и завершениям
func (r *PostgresRepo) GetCommentedVideos(ctx context.Context, limit int) ([]domain.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
		FROM app.videos v
		LEFT JOIN app.video_counters vc ON v.id = vc.video_id
		ORDER BY 
			-- Комбинация лайков и завершений просмотров
			COALESCE(vc.likes, 0) * 0.7 + 
			COALESCE(vc.completes, 0) * 0.3 DESC
		LIMIT $1
	`

	rows, err := r.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// GetRandomVideos возвращает случайную выборку видео
func (r *PostgresRepo) GetRandomVideos(ctx context.Context, limit int) ([]domain.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
		FROM app.videos v
		ORDER BY random()
		LIMIT $1
	`

	rows, err := r.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// GetUserTopTags возвращает топ теги пользователя на основе его активности
func (r *PostgresRepo) GetUserTopTags(ctx context.Context, userID string) ([]string, error) {
	query := `
		WITH user_activity AS (
			SELECT 
				v.tags,
				COUNT(*) as activity_count
			FROM app.events e
			JOIN app.videos v ON e.video_id = v.id
			WHERE e.user_id = $1::uuid
			  AND e.type IN ('view_start', 'like')
			GROUP BY v.tags
		),
		tag_counts AS (
			SELECT 
				unnest(tags) as tag,
				SUM(activity_count) as total_count
			FROM user_activity
			GROUP BY tag
			ORDER BY total_count DESC
		)
		SELECT tag
		FROM tag_counts
		LIMIT 10
	`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetSessionTopTags возвращает топ теги сессии на основе активности
func (r *PostgresRepo) GetSessionTopTags(ctx context.Context, sessionID string) ([]string, error) {
	query := `
		WITH session_activity AS (
			SELECT 
				v.tags,
				COUNT(*) as activity_count
			FROM app.events e
			JOIN app.videos v ON e.video_id = v.id
			WHERE e.session_id = $1
			  AND e.type IN ('view_start', 'like')
			GROUP BY v.tags
		),
		tag_counts AS (
			SELECT 
				unnest(tags) as tag,
				SUM(activity_count) as total_count
			FROM session_activity
			GROUP BY tag
			ORDER BY total_count DESC
		)
		SELECT tag
		FROM tag_counts
		LIMIT 10
	`

	rows, err := r.DB.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetVideosByTags возвращает видео по тегам с релевантностью
func (r *PostgresRepo) GetVideosByTags(ctx context.Context, tags []string, limit int) ([]domain.Video, error) {
	if len(tags) == 0 {
		return []domain.Video{}, nil
	}

	query := `
		SELECT DISTINCT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
		FROM app.videos v
		WHERE v.tags && $1::text[]
		ORDER BY 
			-- Количество совпадающих тегов
			array_length(array(
				SELECT unnest(v.tags) INTERSECT SELECT unnest($1::text[])
			), 1) DESC,
			-- Популярность
			COALESCE((SELECT views FROM app.video_counters vc WHERE vc.video_id = v.id), 0) DESC,
			-- Свежесть
			v.uploaded_at DESC
		LIMIT $2
	`

	rows, err := r.DB.Query(ctx, query, tags, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// GetSimilarVideos возвращает похожие видео на основе тегов
func (r *PostgresRepo) GetSimilarVideos(ctx context.Context, videoID string, limit int) ([]domain.Video, error) {
	query := `
		WITH target_video AS (
			SELECT tags FROM app.videos WHERE id = $1::uuid
		)
		SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
		FROM app.videos v, target_video tv
		WHERE v.id != $1::uuid
		  AND v.tags && tv.tags
		ORDER BY 
			-- Количество общих тегов
			array_length(array(
				SELECT unnest(v.tags) INTERSECT SELECT unnest(tv.tags)
			), 1) DESC,
			-- Популярность
			COALESCE((SELECT views FROM app.video_counters vc WHERE vc.video_id = v.id), 0) DESC
		LIMIT $2
	`

	rows, err := r.DB.Query(ctx, query, videoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// GetDiversifiedVideos возвращает диверсифицированные видео (исключая уже просмотренные)
func (r *PostgresRepo) GetDiversifiedVideos(ctx context.Context, excludeIDs []string, limit int) ([]domain.Video, error) {
	var query string
	var args []interface{}

	if len(excludeIDs) > 0 {
		query = `
			SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
			FROM app.videos v
			WHERE v.id != ALL($1::uuid[])
			ORDER BY 
				-- Случайность для диверсификации
				random(),
				-- Популярность как fallback
				COALESCE((SELECT views FROM app.video_counters vc WHERE vc.video_id = v.id), 0) DESC
			LIMIT $2
		`
		args = []interface{}{excludeIDs, limit}
	} else {
		query = `
			SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id
			FROM app.videos v
			ORDER BY random()
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var video domain.Video
		err := rows.Scan(
			&video.ID,
			&video.Title,
			&video.Description,
			&video.Lang,
			&video.Tags,
			&video.DurationS,
			&video.UploadedAt,
			&video.AuthorID,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (r *PostgresRepo) StatsTotals(ctx context.Context, from, to string) (domain.StatsTotals, error) {
	q := `
		SELECT 
			COALESCE(SUM(views),0),
			COALESCE(SUM(completes),0),
			COALESCE(SUM(likes),0),
			COALESCE(SUM(clicks),0),
			COALESCE(SUM(impressions),0),
			COALESCE(SUM(dwell_ms_sum),0)
		FROM app.video_daily
		WHERE day >= $1::date AND day <= $2::date
	`
	row := r.DB.QueryRow(ctx, q, from, to)
	var t domain.StatsTotals
	if err := row.Scan(&t.Views, &t.Completes, &t.Likes, &t.Clicks, &t.Impressions, &t.DwellMsSum); err != nil {
		return domain.StatsTotals{}, err
	}
	return t, nil
}

func (r *PostgresRepo) StatsTopVideos(ctx context.Context, from, to string, top int) ([]domain.VideoWithStats, error) {
	q := `
		WITH per_video AS (
			SELECT 
				vd.video_id,
				SUM(vd.views) as views,
				SUM(vd.completes) as completes,
				SUM(vd.likes) as likes,
				SUM(vd.clicks) as clicks,
				SUM(vd.impressions) as impressions,
				SUM(vd.views) 
					+ SUM(vd.likes)*2.0
					+ SUM(vd.completes)*1.5
					+ SUM(vd.clicks)*1.2 as score
			FROM app.video_daily vd
			WHERE vd.day >= $1::date AND vd.day <= $2::date
			GROUP BY vd.video_id
		)
		SELECT v.id, v.title, v.description, v.lang, v.tags, v.duration_s, v.uploaded_at, v.author_id,
			p.views, p.completes, p.likes, p.clicks, p.impressions, p.score
		FROM per_video p
		JOIN app.videos v ON v.id = p.video_id
		ORDER BY p.score DESC
		LIMIT $3
	`
	rows, err := r.DB.Query(ctx, q, from, to, top)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.VideoWithStats
	for rows.Next() {
		var vws domain.VideoWithStats
		if err := rows.Scan(
			&vws.Video.ID,
			&vws.Video.Title,
			&vws.Video.Description,
			&vws.Video.Lang,
			&vws.Video.Tags,
			&vws.Video.DurationS,
			&vws.Video.UploadedAt,
			&vws.Video.AuthorID,
			&vws.Views,
			&vws.Completes,
			&vws.Likes,
			&vws.Clicks,
			&vws.Impressions,
			&vws.Score,
		); err != nil {
			return nil, err
		}
		res = append(res, vws)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *PostgresRepo) CreateUser(ctx context.Context, id, email, passHash string) (string, error) {
	var rid string

	err := r.DB.QueryRow(ctx, `
		INSERT INTO app.users(id, email, pass_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`, id, email, passHash).Scan(&rid)
	return rid, err
}

func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (string, string, error) {
	var id, hash string
	err := r.DB.QueryRow(ctx, `
		SELECT id::text, pass_hash FROM app.users WHERE email=$1
	`, email).Scan(&id, &hash)
	return id, hash, err
}
