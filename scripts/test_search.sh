#!/bin/bash

# Скрипт для тестирования ручки /search
# Убедитесь, что API запущен на localhost:8080

API_URL="http://localhost:8080"

echo "🧪 Тестирование ручки /search"
echo "================================"

# Тест 1: Базовый поиск
echo "1. Базовый поиск по запросу 'программирование'"
curl -s "${API_URL}/search?q=go" | jq '.' 2>/dev/null || curl -s "${API_URL}/search?q=go"
echo -e "\n"

# Тест 2: Поиск с ограничением результатов
echo "2. Поиск с limit=5"
curl -s "${API_URL}/search?q=go&limit=5" | jq '.' 2>/dev/null || curl -s "${API_URL}/search?q=go&limit=5"
echo -e "\n"

# Тест 3: Поиск с пагинацией
echo "3. Поиск с offset=0 и limit=3"
curl -s "${API_URL}/search?q=видео&limit=3&offset=0" | jq '.' 2>/dev/null || curl -s "${API_URL}/search?q=видео&limit=3&offset=0"
echo -e "\n"

# Тест 4: Поиск на английском языке
echo "4. Поиск на английском языке"
curl -s "${API_URL}/search?q=programming&limit=10" | jq '.' 2>/dev/null || curl -s "${API_URL}/search?q=programming&limit=10"
echo -e "\n"

# Тест 5: Ошибка - отсутствует обязательный параметр q
echo "5. Тест ошибки - отсутствует параметр 'q'"
curl -s "${API_URL}/search" | jq '.' 2>/dev/null || curl -s "${API_URL}/search"
echo -e "\n"

# Тест 6: Неверные параметры (должны использоваться значения по умолчанию)
echo "6. Тест с неверными параметрами (limit=invalid, offset=invalid)"
curl -s "${API_URL}/search?q=test&limit=invalid&offset=invalid" | jq '.' 2>/dev/null || curl -s "${API_URL}/search?q=test&limit=invalid&offset=invalid"
echo -e "\n"

echo "✅ Тестирование завершено!"
echo "Примечание: Если API не запущен, вы увидите ошибки соединения"
echo "Для запуска API выполните: go run ./cmd/api" 