# Swagger UI Setup для MicroTube API

## 🎯 Что добавлено

1. **Swagger UI интерфейс** - доступен по адресу `/docs`
2. **Автоматическое перенаправление** с корневой страницы `/` на `/docs`
3. **OpenAPI спецификация** - доступна по адресу `/openapi.yaml`
4. **Автоматический seed данных** при запуске через Docker Compose

## 🚀 Быстрый запуск

### 1. Запуск всех сервисов

```bash
make docker-up
```

### 2. Проверка статуса

```bash
make docker-logs
```

### 3. Открытие Swagger UI в браузере

Перейдите по адресу: **http://localhost:8080/docs**

## 📊 Доступные эндпоинты

- **GET /** - перенаправление на документацию
- **GET /docs** - Swagger UI интерфейс
- **GET /openapi.yaml** - OpenAPI спецификация в формате YAML
- **GET /healthz** - проверка здоровья сервиса

## 🔧 API Endpoints

### Аутентификация

- `POST /auth/register` - регистрация пользователя
- `POST /auth/login` - вход пользователя

### События

- `POST /events` - отправка событий (идемпотентно)

### Поиск и контент

- `GET /search` - поиск видео (FTS + триграммы)
- `GET /videos/feed` - лента видео (популярные, комментируемые, случайные)
- `GET /recommendations` - персонализированные рекомендации

### Аналитика

- `GET /stats/overview` - статистика (только для админов)

## 🐳 Docker Compose команды

```bash
# Основные команды
make docker-up          # Запустить все сервисы
make docker-down        # Остановить все сервисы
make docker-build       # Пересобрать и запустить API
make docker-logs        # Логи всех сервисов
make docker-api-logs    # Логи только API
make docker-seed        # Запустить только seed
make docker-reseed      # Перезапустить seed с пересборкой
```

## 🔍 Проверка работоспособности

### 1. Health Check

```bash
curl http://localhost:8080/healthz
```

**Ожидаемый ответ:** `{"status":"ok"}`

### 2. Swagger UI

Откройте http://localhost:8080/docs в браузере

### 3. OpenAPI спецификация

```bash
curl http://localhost:8080/openapi.yaml
```

## 🎨 Особенности Swagger UI

- **Интерактивная документация** - тестируйте API прямо из браузера
- **Авторизация** - введите JWT токен для защищенных эндпоинтов
- **Примеры запросов** - готовые примеры для каждого эндпоинта
- **Схемы данных** - полное описание структур запросов и ответов

## 🚨 Устранение неполадок

### Swagger UI не загружается

1. Проверьте, что API сервер запущен: `docker compose ps`
2. Проверьте логи API: `make docker-api-logs`
3. Убедитесь, что порт 8080 свободен

### Seed не выполняется

1. Проверьте логи seed: `make docker-seed-logs`
2. Перезапустите seed: `make docker-reseed`

### База данных недоступна

1. Проверьте статус PostgreSQL: `docker compose ps db`
2. Проверьте логи базы: `docker compose logs db`

## 📚 Дополнительные ресурсы

- **Основной README**: [README.md](./README.md)
- **Docker Compose**: [docker-compose.yml](./docker-compose.yml)
- **Makefile**: [Makefile](./Makefile)

## 🎯 Преимущества настройки

1. **Автоматизация** - все сервисы запускаются одной командой
2. **Документация** - Swagger UI всегда актуален
3. **Тестирование** - удобное тестирование API через браузер
4. **Разработка** - быстрый старт для новых разработчиков
5. **Демонстрация** - идеально для собеседований и презентаций
