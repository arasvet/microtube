SET search_path TO app, public;

-- Обновляем tsvector для полнотекстового поиска
UPDATE videos
SET fts_tsv =
        setweight(to_tsvector('simple', immutable_unaccent(coalesce(title,''))), 'A')
            || setweight(to_tsvector('simple', immutable_unaccent(coalesce(description,''))), 'B');

-- Индексы для поиска и сортировки
CREATE INDEX IF NOT EXISTS videos_fts_idx
    ON videos USING GIN (fts_tsv);

CREATE INDEX IF NOT EXISTS videos_title_trgm_idx
    ON videos USING GIN ( (immutable_unaccent(title)) gin_trgm_ops );

CREATE INDEX IF NOT EXISTS videos_desc_trgm_idx
    ON videos USING GIN ( (immutable_unaccent(description)) gin_trgm_ops );

CREATE INDEX IF NOT EXISTS videos_uploaded_at_idx
    ON videos (uploaded_at);
