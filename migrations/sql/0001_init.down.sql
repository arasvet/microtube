SET search_path TO app, public;

DROP TABLE IF EXISTS user_signals;
DROP TABLE IF EXISTS video_daily;
DROP TABLE IF EXISTS video_counters;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS event_type;

DROP FUNCTION IF EXISTS immutable_unaccent(text);

DROP SCHEMA IF EXISTS app CASCADE;
