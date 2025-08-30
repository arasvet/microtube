-- Улучшение данных для тестирования фидов
SET search_path TO app,
    public;
-- 1. Создаем несколько видео с высокими показателями для популярного фида
INSERT INTO app.videos (
        id,
        title,
        description,
        lang,
        tags,
        duration_s,
        uploaded_at
    )
VALUES -- Видео с высокими просмотрами и лайками (недавнее)
    (
        gen_random_uuid(),
        'Go Microservices Masterclass',
        'Complete guide to building microservices with Go',
        'en',
        '{go, microservices, backend}',
        1800,
        now() - interval '2 days'
    ),
    (
        gen_random_uuid(),
        'Docker for Beginners',
        'Learn Docker from scratch with practical examples',
        'en',
        '{docker, devops, cloud}',
        1200,
        now() - interval '5 days'
    ),
    (
        gen_random_uuid(),
        'PostgreSQL Performance Tuning',
        'Advanced techniques for optimizing PostgreSQL',
        'en',
        '{postgres, database, performance}',
        2400,
        now() - interval '1 week'
    ),
    -- Видео с высокими просмотрами, но старые (проверим затухание по времени)
    (
        gen_random_uuid(),
        'Redis Caching Strategies',
        'Best practices for Redis implementation',
        'en',
        '{redis, caching, performance}',
        900,
        now() - interval '3 months'
    ),
    (
        gen_random_uuid(),
        'Kafka Stream Processing',
        'Real-time data processing with Apache Kafka',
        'en',
        '{kafka, streaming, big-data}',
        1500,
        now() - interval '6 months'
    ),
    -- Видео с низкими показателями (для сравнения)
    (
        gen_random_uuid(),
        'Introduction to Testing',
        'Basic testing concepts in software development',
        'en',
        '{testing, quality, basics}',
        600,
        now() - interval '2 weeks'
    ),
    (
        gen_random_uuid(),
        'Frontend Fundamentals',
        'HTML, CSS, and JavaScript basics',
        'en',
        '{frontend, web, basics}',
        800,
        now() - interval '1 month'
    );
-- 2. Добавляем счетчики для популярного фида
-- Высокие показатели для недавних видео
INSERT INTO app.video_counters (video_id, views, completes, likes, last_event_at)
SELECT v.id,
    CASE
        WHEN v.title LIKE '%Masterclass%' THEN 15000
        WHEN v.title LIKE '%Docker%' THEN 12000
        WHEN v.title LIKE '%PostgreSQL%' THEN 18000
        ELSE 1000
    END as views,
    CASE
        WHEN v.title LIKE '%Masterclass%' THEN 8000
        WHEN v.title LIKE '%Docker%' THEN 6000
        WHEN v.title LIKE '%PostgreSQL%' THEN 10000
        ELSE 500
    END as completes,
    CASE
        WHEN v.title LIKE '%Masterclass%' THEN 1200
        WHEN v.title LIKE '%Docker%' THEN 900
        WHEN v.title LIKE '%PostgreSQL%' THEN 1500
        ELSE 100
    END as likes,
    v.uploaded_at
FROM app.videos v
WHERE v.title IN (
        'Go Microservices Masterclass',
        'Docker for Beginners',
        'PostgreSQL Performance Tuning'
    ) ON CONFLICT (video_id) DO
UPDATE
SET views = EXCLUDED.views,
    completes = EXCLUDED.completes,
    likes = EXCLUDED.likes,
    last_event_at = EXCLUDED.last_event_at;
-- Средние показатели для старых видео
INSERT INTO app.video_counters (video_id, views, completes, likes, last_event_at)
SELECT v.id,
    CASE
        WHEN v.title LIKE '%Redis%' THEN 8000
        WHEN v.title LIKE '%Kafka%' THEN 10000
        ELSE 1000
    END as views,
    CASE
        WHEN v.title LIKE '%Redis%' THEN 4000
        WHEN v.title LIKE '%Kafka%' THEN 5000
        ELSE 500
    END as completes,
    CASE
        WHEN v.title LIKE '%Redis%' THEN 600
        WHEN v.title LIKE '%Kafka%' THEN 800
        ELSE 100
    END as likes,
    v.uploaded_at
FROM app.videos v
WHERE v.title IN (
        'Redis Caching Strategies',
        'Kafka Stream Processing'
    ) ON CONFLICT (video_id) DO
UPDATE
SET views = EXCLUDED.views,
    completes = EXCLUDED.completes,
    likes = EXCLUDED.likes,
    last_event_at = EXCLUDED.last_event_at;
-- 3. Создаем видео с высокими лайками для комментируемого фида
INSERT INTO app.videos (
        id,
        title,
        description,
        lang,
        tags,
        duration_s,
        uploaded_at
    )
VALUES (
        gen_random_uuid(),
        'AI in Modern Applications',
        'How to integrate AI into your software',
        'en',
        '{ai, ml, integration}',
        1600,
        now() - interval '1 week'
    ),
    (
        gen_random_uuid(),
        'DevOps Best Practices',
        'Industry standards for DevOps implementation',
        'en',
        '{devops, best-practices, automation}',
        1400,
        now() - interval '2 weeks'
    ),
    (
        gen_random_uuid(),
        'Cloud Architecture Patterns',
        'Designing scalable cloud applications',
        'en',
        '{cloud, architecture, scalability}',
        2000,
        now() - interval '3 weeks'
    );
-- Высокие лайки, средние завершения для комментируемого фида
INSERT INTO app.video_counters (video_id, views, completes, likes, last_event_at)
SELECT v.id,
    CASE
        WHEN v.title LIKE '%AI%' THEN 5000
        WHEN v.title LIKE '%DevOps%' THEN 6000
        WHEN v.title LIKE '%Cloud%' THEN 7000
        ELSE 1000
    END as views,
    CASE
        WHEN v.title LIKE '%AI%' THEN 2500
        WHEN v.title LIKE '%DevOps%' THEN 3000
        WHEN v.title LIKE '%Cloud%' THEN 3500
        ELSE 500
    END as completes,
    CASE
        WHEN v.title LIKE '%AI%' THEN 800
        WHEN v.title LIKE '%DevOps%' THEN 1000
        WHEN v.title LIKE '%Cloud%' THEN 1200
        ELSE 100
    END as likes,
    v.uploaded_at
FROM app.videos v
WHERE v.title IN (
        'AI in Modern Applications',
        'DevOps Best Practices',
        'Cloud Architecture Patterns'
    ) ON CONFLICT (video_id) DO
UPDATE
SET views = EXCLUDED.views,
    completes = EXCLUDED.completes,
    likes = EXCLUDED.likes,
    last_event_at = EXCLUDED.last_event_at;
-- 4. Обновляем FTS индексы для новых видео
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
-- 5. Показываем статистику
SELECT 'Enhanced feeds data added' as status,
    COUNT(*) as total_videos,
    COUNT(vc.video_id) as videos_with_counters
FROM app.videos v
    LEFT JOIN app.video_counters vc ON v.id = vc.video_id;