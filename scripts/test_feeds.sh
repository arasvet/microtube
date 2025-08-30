#!/bin/bash

# Скрипт для тестирования ручки /videos/feed
# Убедитесь, что API запущен на localhost:8080

API_URL="http://localhost:8080"

echo "🧪 Тестирование ручки /videos/feed"
echo "=================================="

# Тест 1: Популярные видео (по умолчанию)
echo "1. Популярные видео (по умолчанию)"
curl -s "${API_URL}/videos/feed" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed"
echo -e "\n"

# Тест 2: Популярные видео с ограничением
echo "2. Популярные видео с limit=5"
curl -s "${API_URL}/videos/feed?type=popular&limit=5" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed?type=popular&limit=5"
echo -e "\n"

# Тест 3: Комментируемые видео
echo "3. Комментируемые видео с limit=3"
curl -s "${API_URL}/videos/feed?type=commented&limit=3" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed?type=commented&limit=3"
echo -e "\n"

# Тест 4: Случайные видео
echo "4. Случайные видео с limit=3"
curl -s "${API_URL}/videos/feed?type=random&limit=3" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed?type=random&limit=3"
echo -e "\n"

# Тест 5: Неверный тип (должен вернуть популярные по умолчанию)
echo "5. Неверный тип (должен вернуть популярные по умолчанию)"
curl -s "${API_URL}/videos/feed?type=invalid" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed?type=invalid"
echo -e "\n"

# Тест 6: Неверный limit (должен использоваться по умолчанию)
echo "6. Неверный limit (должен использоваться по умолчанию)"
curl -s "${API_URL}/videos/feed?type=popular&limit=invalid" | jq '.type, .limit, .total' 2>/dev/null || curl -s "${API_URL}/videos/feed?type=popular&limit=invalid"
echo -e "\n"

echo "✅ Тестирование фидов завершено!"
echo "Примечание: Если API не запущен, вы увидите ошибки соединения"
echo "Для запуска API выполните: go run ./cmd/api" 