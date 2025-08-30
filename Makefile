PROJECT ?= microtube
POSTGRES_USER ?= app
POSTGRES_PASSWORD ?= app
POSTGRES_DB ?= microtube
DB_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: up down logs psql db-shell redis-cli migrate-up migrate-down migrate-new

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

