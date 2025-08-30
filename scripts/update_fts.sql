-- Обновление FTS индексов для всех видео
SET search_path TO app,
    public;
-- Обновляем tsvector для всех видео
UPDATE videos
SET fts_tsv = setweight(
        to_tsvector(
            'simple',
            immutable_unaccent(coalesce(title, ''))
        ),
        'A'
    ) || setweight(
        to_tsvector(
            'simple',
            immutable_unaccent(coalesce(description, ''))
        ),
        'B'
    )
WHERE fts_tsv IS NULL
    OR fts_tsv = ''::tsvector;
-- Проверяем результат
SELECT 'FTS update completed' as status,
    COUNT(*) as total_videos,
    COUNT(fts_tsv) as videos_with_fts
FROM videos;