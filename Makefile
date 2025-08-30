PROJECT ?= microtube
POSTGRES_USER ?= app
POSTGRES_PASSWORD ?= app
POSTGRES_DB ?= microtube
DB_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: up down logs psql db-shell redis-cli migrate-up migrate-down migrate-new docker-up docker-down docker-build docker-logs

up: ## старт БД и Редиса
	docker compose up -d db redis

down: ## остановка и очистка (включая volume'ы)
	docker compose down -v

logs: ## логи сервисов
	docker compose logs -f --tail=200

psql: ## открыть psql в контейнере Postgres
	docker exec -it microtube-db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-shell: ## sh внутри контейнера Postgres
	docker exec -it microtube-db sh

redis-cli: ## redis-cli в контейнере
	docker exec -it microtube-redis redis-cli

migrate-up: ## применить все миграции
	docker run --rm --network host -v $$PWD/migrations/sql:/migrations \
		migrate/migrate:4 \
		-path=/migrations -database "$(DB_URL)" up

migrate-down: ## откатить одну миграцию
	docker run --rm --network host -v $$PWD/migrations/sql:/migrations \
		migrate/migrate:4 \
		-path=/migrations -database "$(DB_URL)" down 1

seed: ## прогнать сидер (можно VIDEOS=... EVENTS=...)
	export $(shell grep -v '^#' .env | xargs) && go run ./cmd/seed

run: ## запустить API локально
	export $(shell grep -v '^#' .env | xargs) && go run ./cmd/api

# Docker Compose команды
docker-up: ## запустить все сервисы через Docker Compose
	docker compose up -d

docker-down: ## остановить все сервисы Docker Compose
	docker compose down

docker-build: ## пересобрать и запустить API
	docker compose up -d --build api

docker-seed: ## запустить только seed
	docker compose up seed

docker-reseed: ## перезапустить seed с пересборкой
	docker compose up --build seed

docker-logs: ## логи всех сервисов Docker Compose
	docker compose logs -f --tail=200

docker-api-logs: ## логи только API сервиса
	docker compose logs -f --tail=200 api

docker-seed-logs: ## логи только seed сервиса
	docker compose logs -f --tail=200 seed

# Тестирование
test: ## запустить все тесты
	go test ./... -v

test-http: ## запустить тесты HTTP handlers
	go test ./internal/http/... -v

test-coverage: ## запустить тесты с покрытием
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

