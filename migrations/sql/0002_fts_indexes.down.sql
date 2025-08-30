SET search_path TO app, public;

DROP INDEX IF EXISTS videos_uploaded_at_idx;
DROP INDEX IF EXISTS videos_desc_trgm_idx;
DROP INDEX IF EXISTS videos_title_trgm_idx;
DROP INDEX IF EXISTS videos_fts_idx;

ALTER TABLE videos DROP COLUMN IF EXISTS fts_tsv;
