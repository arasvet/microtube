#!/bin/bash

# Скрипт для улучшения данных фидов и их тестирования

echo "🚀 Улучшение данных для тестирования фидов"
echo "=========================================="

# 1. Добавляем улучшенные данные
echo "1. Добавляем улучшенные данные для фидов..."
docker exec -i microtube-db psql -U app -d microtube < scripts/enhance_feeds.sql

echo -e "\n2. Проверяем популярные видео (должны быть новые высокорейтинговые видео):"
echo "=== Популярные видео (top 5) ==="
curl -s "http://localhost:8080/videos/feed?type=popular&limit=5" | jq '.videos[] | {title: .Title, duration: .DurationS, uploaded: .UploadedAt}' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=popular&limit=5"

echo -e "\n3. Проверяем комментируемые видео (должны быть видео с высокими лайками):"
echo "=== Комментируемые видео (top 5) ==="
curl -s "http://localhost:8080/videos/feed?type=commented&limit=5" | jq '.videos[] | {title: .Title, duration: .DurationS, uploaded: .UploadedAt}' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=commented&limit=5"

echo -e "\n4. Проверяем случайные видео (должны быть разные при каждом запросе):"
echo "=== Случайные видео (запрос 1) ==="
curl -s "http://localhost:8080/videos/feed?type=random&limit=3" | jq '.videos[] | {title: .Title}' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=random&limit=3"

echo -e "\n=== Случайные видео (запрос 2) ==="
curl -s "http://localhost:8080/videos/feed?type=random&limit=3" | jq '.videos[] | {title: .Title}' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=random&limit=3"

echo -e "\n5. Анализируем логику популярного фида:"
echo "=== Сравнение: новые vs старые видео ==="
echo "Новые видео должны быть выше из-за затухания по времени:"
curl -s "http://localhost:8080/videos/feed?type=popular&limit=10" | jq '.videos[] | {title: .Title, uploaded: .UploadedAt}' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=popular&limit=10"

echo -e "\n6. Тестируем граничные случаи:"
echo "=== Очень большой limit ==="
curl -s "http://localhost:8080/videos/feed?type=popular&limit=999" | jq '.limit, .total' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=popular&limit=999"

echo -e "\n=== Отрицательный limit ==="
curl -s "http://localhost:8080/videos/feed?type=popular&limit=-5" | jq '.limit, .total' 2>/dev/null || curl -s "http://localhost:8080/videos/feed?type=popular&limit=-5"

echo -e "\n✅ Тестирование улучшенных фидов завершено!"
echo "Теперь у нас есть более разнообразные данные для тестирования логики ранжирования." 