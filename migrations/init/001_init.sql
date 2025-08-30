-- Создаём расширения в public (стандартное место)
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;
-- Создаём отдельную схему для нашего приложения
CREATE SCHEMA IF NOT EXISTS app AUTHORIZATION CURRENT_USER;
SET search_path TO app,
    public;
-- Функция-обёртка, чтобы unaccent был IMMUTABLE (можно использовать в индексах)
CREATE OR REPLACE FUNCTION immutable_unaccent(text) RETURNS text AS $$
SELECT unaccent('public.unaccent', $1);
$$ LANGUAGE sql IMMUTABLE;
-- Таблицы
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    email text UNIQUE NOT NULL,
    pass_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS videos (
    id uuid PRIMARY KEY,
    title text NOT NULL,
    description text NOT NULL,
    lang text NOT NULL,
    tags text [] NOT NULL DEFAULT '{}',
    duration_s int NOT NULL,
    uploaded_at timestamptz NOT NULL,
    author_id uuid,
    CONSTRAINT videos_author_fk FOREIGN KEY (author_id) REFERENCES users(id)
);
-- Поле для полнотекстового поиска
ALTER TABLE videos
ADD COLUMN IF NOT EXISTS fts_tsv tsvector;
-- Тип событий
DO $$ BEGIN CREATE TYPE event_type AS ENUM (
    'view_start',
    'view_complete',
    'like',
    'search_query',
    'click_result'
);
EXCEPTION
WHEN duplicate_object THEN null;
END $$;
-- События
CREATE TABLE IF NOT EXISTS events (
    event_id uuid PRIMARY KEY,
    ts timestamptz NOT NULL,
    type event_type NOT NULL,
    session_id text NOT NULL,
    user_id uuid,
    video_id uuid,
    query text,
    dwell_ms int,
    CONSTRAINT events_video_fk FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE
    SET NULL,
        CONSTRAINT events_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE
    SET NULL
);
-- Денорм. счётчики
CREATE TABLE IF NOT EXISTS video_counters (
    video_id uuid PRIMARY KEY REFERENCES videos(id),
    views bigint NOT NULL DEFAULT 0,
    completes bigint NOT NULL DEFAULT 0,
    likes bigint NOT NULL DEFAULT 0,
    comments bigint NOT NULL DEFAULT 0,
    last_event_at timestamptz
);
-- Суточная аналитика
CREATE TABLE IF NOT EXISTS video_daily (
    video_id uuid NOT NULL REFERENCES videos(id),
    day date NOT NULL,
    views bigint NOT NULL DEFAULT 0,
    completes bigint NOT NULL DEFAULT 0,
    likes bigint NOT NULL DEFAULT 0,
    clicks bigint NOT NULL DEFAULT 0,
    impressions bigint NOT NULL DEFAULT 0,
    dwell_ms_sum bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (video_id, day)
);
-- Сигналы пользователя/сессии
CREATE TABLE IF NOT EXISTS user_signals (
    user_or_session text PRIMARY KEY,
    top_tags text [] NOT NULL DEFAULT '{}',
    last_seen_at timestamptz
);