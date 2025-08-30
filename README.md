# MicroTube - Микросервис для рекомендаций видео

Микросервис для сбора событий пользователей, поиска видео и персонализированных рекомендаций.

## 🚀 Быстрый старт (Docker Compose)

### Требования

- Docker Desktop
- Docker Compose
- `jq` для форматирования JSON (опционально)
- `python3` для извлечения user_id из JWT (опционально)

**Установка дополнительных инструментов:**

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq python3

# CentOS/RHEL
sudo yum install jq python3
```

### Пошаговая инструкция

1. **Клонирование репозитория**

```bash
git clone https://github.com/arasvet/microtube.git
cd microtube
```

2. **Запуск всех сервисов**

```bash
make docker-up
```

3. **Проверка статуса**

```bash
make docker-logs
```

4. **Открытие Swagger UI в браузере**

```
http://localhost:8080/docs
```

**Что происходит при запуске:**

- PostgreSQL и Redis запускаются
- Автоматически выполняется seed данных (1200 видео, 50000 событий)
- API сервер запускается после успешного seed
- Swagger UI становится доступен для тестирования

## 📊 Проверка работоспособности

### 1. Health Check

```bash
curl http://localhost:8080/healthz
```

**Ожидаемый ответ:** `{"status":"ok"}`

### 2. Проверка Swagger UI

Откройте в браузере: http://localhost:8080/docs

Вы должны увидеть:

- Полную документацию API
- Все доступные эндпоинты
- Возможность тестировать API прямо из браузера

### 3. Проверка OpenAPI спецификации

```bash
curl http://localhost:8080/openapi.yaml
```

**Ожидаемый ответ:** YAML файл с OpenAPI спецификацией

## 🔧 Доступные команды Make

```bash
# Основные команды
make docker-up          # Запустить все сервисы
make docker-down        # Остановить все сервисы
make docker-build       # Пересобрать и запустить API
make docker-logs        # Логи всех сервисов
make docker-api-logs    # Логи только API
make docker-seed        # Запустить только seed
make docker-reseed      # Перезапустить seed с пересборкой

# Тестирование
make test               # Запустить все тесты
make test-http          # Запустить тесты HTTP handlers
make test-coverage      # Запустить тесты с покрытием
```

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Только HTTP handlers
make test-http

# Тесты с покрытием
make test-coverage
```

### Покрытие тестами

Проект включает тесты для:

- **HTTP Handlers** - тестирование всех API эндпоинтов
- **Mock объекты** - изолированное тестирование бизнес-логики
- **Edge cases** - проверка граничных случаев и ошибок

## 🧪 Тестирование API

### Полное тестирование API

Вот полный набор команд для тестирования всех возможностей API:

#### 1. Регистрация и авторизация

```bash
# Регистрация пользователя
curl -i -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"password123"}'

# Логин и получение токена
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"password123"}')

echo "$LOGIN_RESPONSE" | jq .
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .token)
echo "TOKEN: ${#TOKEN} chars"
```

#### 2. Извлечение user_id из JWT

```bash
USER_ID=$(python3 -c "
import base64, json
parts = '$TOKEN'.split('.')
payload = parts[1] + '=' * ((4 - len(parts[1]) % 4) % 4)
print(json.loads(base64.urlsafe_b64decode(payload))['sub'])
")
echo "USER_ID=$USER_ID"
```

#### 3. Тестирование событий (идемпотентность)

```bash
# Получить ID видео из базы
VIDEO_ID=$(docker compose exec db psql -U app -d microtube -t -c "select id from app.videos limit 1" | tr -d '[:space:]')
echo "VIDEO_ID: $VIDEO_ID"

# Отправить событие
curl -i -X POST http://localhost:8080/events \
  -H 'Content-Type: application/json' \
  -d "{
    \"event_id\": \"test-event-$(date +%s)\",
    \"ts\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"type\": \"view_start\",
    \"session_id\": \"sess-1\",
    \"video_id\": \"$VIDEO_ID\"
  }"
```

#### 4. Тестирование поиска

```bash
# Поиск по ключевому слову
curl -s "http://localhost:8080/search?q=go&limit=5" | jq '.results[0:3]'

# Поиск с пагинацией
curl -s "http://localhost:8080/search?q=test&limit=3&offset=0" | jq '.total,.limit,.offset'
```

#### 5. Тестирование фидов

```bash
# Популярные видео
curl -s "http://localhost:8080/videos/feed?type=popular&limit=5" | jq '.videos[0:3]'

# Комментируемые видео
curl -s "http://localhost:8080/videos/feed?type=commented&limit=3" | jq '.videos[0:3]'

# Случайные видео
curl -s "http://localhost:8080/videos/feed?type=random&limit=3" | jq '.videos[0:3]'
```

#### 6. Тестирование рекомендаций

```bash
# Холодные рекомендации (для гостей)
curl -s "http://localhost:8080/recommendations?session_id=test-session&limit=5" | jq '.type,.total'

# Персональные рекомендации (для авторизованных)
curl -s "http://localhost:8080/recommendations?user_id=$USER_ID&limit=5" | jq '.type,.total'
```

#### 7. Тестирование статистики (только для админов)

```bash
# Без авторизации -> 403
curl -i "http://localhost:8080/stats/overview?top=3" | head -5

# С JWT токеном
curl -s "http://localhost:8080/stats/overview?top=3" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

### События (идемпотентность)

```bash
# Невалидное тело -> 422
curl -i -X POST http://localhost:8080/events \
  -H 'Content-Type: application/json' \
  -d '{}'

# Валидное событие
curl -i -X POST http://localhost:8080/events \
  -H 'Content-Type: application/json' \
  -d '{
    "event_id": "test-event-123",
    "ts": "2024-01-01T12:00:00Z",
    "type": "view_start",
    "session_id": "sess-1",
    "video_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### Статистика (только для админов)

```bash
# Без авторизации -> 403
curl -i "http://localhost:8080/stats/overview?top=3"

# С JWT токеном (замените TOKEN на полученный)
curl -i "http://localhost:8080/stats/overview?top=3" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Получение user_id из JWT

```bash
# После логина получите TOKEN и извлеките user_id
TOKEN="ваш_jwt_токен_здесь"
USER_ID=$(python3 -c "
import base64, json
parts = '$TOKEN'.split('.')
payload = parts[1] + '=' * ((4 - len(parts[1]) % 4) % 4)
print(json.loads(base64.urlsafe_b64decode(payload))['sub'])
")
echo "USER_ID=$USER_ID"
```

### Проверка FTS-триггера

```bash
# Проверить, что FTS работает корректно
docker compose exec db psql -U app -d microtube -c "
INSERT INTO app.videos(id,title,description,lang,tags,duration_s,uploaded_at)
VALUES (gen_random_uuid(),'FTS Trigger Check','Some text','en','{test}',10,now());
SELECT fts_tsv IS NOT NULL as has_fts FROM app.videos WHERE title='FTS Trigger Check';
"
```

## 🗄️ Структура базы данных

После seed в базе данных будет:

- **1200 видео** с тегами, описаниями и метаданными
- **50000 событий** различных типов (просмотры, лайки, поиски)
- **10 тестовых пользователей** для демонстрации

## 🔍 Основные возможности

### Поиск (Full-Text Search)

- Поиск по названию и описанию видео
- Поддержка русского и английского языков
- Триграмный поиск для опечаток

### Рекомендации

- Холодные рекомендации для новых пользователей
- Персональные рекомендации на основе истории
- Гибридный подход с машинным обучением

### Аналитика

- Сбор событий пользователей
- Статистика просмотров и взаимодействий
- Административная панель для аналитики

## 🐛 Устранение неполадок

### Seed не выполняется

```bash
# Проверить логи seed
make docker-seed-logs

# Перезапустить seed
make docker-reseed
```

### API не отвечает

```bash
# Проверить статус контейнеров
docker compose ps

# Проверить логи API
make docker-api-logs
```

### База данных недоступна

```bash
# Проверить здоровье PostgreSQL
docker compose exec db pg_isready -U app -d microtube
```

## 📁 Структура проекта

```
microtube/
├── cmd/
│   ├── api/          # Основной API сервер
│   └── seed/         # Скрипт заполнения тестовыми данными
├── internal/
│   ├── config/       # Конфигурация
│   ├── domain/       # Доменные модели
│   ├── http/         # HTTP handlers и роутинг
│   ├── repo/         # Репозитории для работы с БД
│   └── usecase/      # Бизнес-логика
├── migrations/        # Миграции базы данных
├── scripts/          # Вспомогательные скрипты
├── docker-compose.yml # Конфигурация Docker Compose
├── Dockerfile        # Dockerfile для API
├── Dockerfile.seed   # Dockerfile для seed
└── Makefile          # Команды для управления
```

## 🌟 Особенности реализации

- **Идемпотентность событий** - повторная отправка события не создает дубли
- **JWT аутентификация** - безопасная авторизация пользователей
- **Graceful shutdown** - корректное завершение работы сервера
- **Health checks** - мониторинг состояния сервисов
- **Swagger UI** - интерактивная документация API
- **Docker Compose** - простое развертывание всех компонентов

## 📞 Поддержка

Если у вас возникли вопросы или проблемы:

1. Проверьте логи: `make docker-logs`
2. Убедитесь, что все контейнеры запущены: `docker compose ps`
3. Проверьте, что порт 8080 свободен
4. Убедитесь, что Docker Desktop запущен

## 🎯 Что демонстрирует проект

- **Архитектура микросервисов** с четким разделением ответственности
- **Работа с PostgreSQL** (миграции, FTS, триггеры)
- **Redis** для кеширования и идемпотентности
- **REST API** с полной документацией
- **Docker** для контейнеризации и развертывания
- **Go** как современный язык для backend разработки
