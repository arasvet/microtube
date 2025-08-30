-- Проверка состояния базы данных
SET search_path TO app,
    public;
-- Проверяем количество видео
SELECT 'Videos count:' as info,
    COUNT(*) as count
FROM videos;
-- Проверяем, есть ли FTS индексы
SELECT 'FTS indexes:' as info,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'videos'
    AND indexname LIKE '%fts%';
-- Проверяем, есть ли trigram индексы
SELECT 'Trigram indexes:' as info,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'videos'
    AND indexname LIKE '%trgm%';
-- Проверяем, есть ли данные в fts_tsv
SELECT 'FTS data check:' as info,
    COUNT(*) as total_videos,
    COUNT(fts_tsv) as videos_with_fts,
    COUNT(*) - COUNT(fts_tsv) as videos_without_fts
FROM videos;
-- Показываем пример видео
SELECT 'Sample video:' as info,
    id,
    title,
    description,
    fts_tsv IS NOT NULL as has_fts
FROM videos
LIMIT 3;