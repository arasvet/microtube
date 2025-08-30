#!/bin/bash

# Скрипт для исправления поиска через обновление FTS индексов

echo "🔧 Исправление поиска в microtube"
echo "=================================="

# Проверяем состояние базы данных
echo "1. Проверяем состояние базы данных..."
docker exec -i microtube-db psql -U app -d microtube < scripts/check_db.sql

echo -e "\n2. Обновляем FTS индексы..."
docker exec -i microtube-db psql -U app -d microtube < scripts/update_fts.sql

echo -e "\n3. Проверяем результат..."
docker exec -i microtube-db psql -U app -d microtube < scripts/check_db.sql

echo -e "\n4. Тестируем поиск..."
echo "Поиск по 'go':"
curl -s "http://localhost:8080/search?q=go" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/search?q=go"

echo -e "\nПоиск по 'Video':"
curl -s "http://localhost:8080/search?q=Video" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/search?q=Video"

echo -e "\n✅ Исправление завершено!" 