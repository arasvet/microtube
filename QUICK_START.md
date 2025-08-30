# 🚀 Quick Start для собеседующего

## ⚡ Быстрый запуск (5 минут)

### 1. Запуск проекта

```bash
git clone https://github.com/arasvet/microtube.git
cd microtube
make docker-up
```

### 2. Проверка работоспособности

```bash
# Проверка статуса
make docker-logs

# Health check
curl http://localhost:8080/healthz

# Открыть Swagger UI в браузере
open http://localhost:8080/docs
```

### 3. Запуск тестов

```bash
make test
```

## 🎯 Что демонстрирует проект

- **Микросервисная архитектура** с четким разделением ответственности
- **Docker Compose** для простого развертывания
- **Swagger UI** для интерактивной документации API
- **Автоматический seed** тестовых данных (1200 видео, 50000 событий)
- **Полное покрытие тестами** HTTP handlers
- **PostgreSQL + Redis** для хранения данных и кеширования
- **Полный набор curl команд** для тестирования всех API эндпоинтов

## 🌐 Доступные адреса

- **Swagger UI**: http://localhost:8080/docs
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/healthz

## 🔧 Основные команды

```bash
make docker-up          # Запустить все сервисы
make docker-down        # Остановить все сервисы
make docker-logs        # Логи всех сервисов
make test               # Запустить все тесты
```

## 📊 API Endpoints

- `POST /auth/register` - регистрация
- `POST /auth/login` - авторизация
- `GET /search` - поиск видео
- `GET /videos/feed` - лента видео
- `GET /recommendations` - рекомендации
- `POST /events` - события пользователей

## 🧪 Тестирование

### Автоматические тесты

Проект включает **12 тестовых кейсов**:

- Поиск видео (5 тестов)
- Лента видео (6 тестов)
- Рекомендации (6 тестов)

Все тесты проходят успешно ✅

### Ручное тестирование API

```bash
# Health check
curl http://localhost:8080/healthz

# Регистрация пользователя
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'

# Поиск видео
curl "http://localhost:8080/search?q=go&limit=5"

# Лента видео
curl "http://localhost:8080/videos/feed?type=popular&limit=5"
```

**Полный набор curl команд** доступен в [README.md](./README.md)

## 🎉 Результат

После выполнения инструкции у вас будет:

- ✅ Работающий API сервер
- ✅ База данных с тестовыми данными
- ✅ Swagger UI для тестирования API
- ✅ Все тесты проходят
- ✅ Готовый к демонстрации проект

**Время выполнения**: ~5 минут
**Сложность**: Минимальная (только Docker)
