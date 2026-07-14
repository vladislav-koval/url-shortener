# Для Windows используем Bash из Git, для других ОС — системный путь
ifeq ($(OS),Windows_NT)
    SHELL := C:/Program Files/Git/bin/bash.exe
else
    SHELL := bash
endif

include .env
export

run:
	@go mod tidy && \
	go run cmd/urlshortener/main.go

env-up:
	@docker-compose up -d

env-down:
	@docker-compose down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Parameter name is required. Example: make migrate-create name=init"; \
		exit 1; \
	fi; \

	docker-compose run --rm postgres-migrate \
		create \
		-ext sql \
		-dir ./migrations \
		-seq $(name)

env-port-forward:
	@docker-compose up -d port-forwarder

env-port-close:
	@docker-compose down port-forwarder

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Parameter action is required. Example: make migrate-action action=up 1"; \
		exit 1; \
	fi; \

	docker-compose run --rm postgres-migrate \
		-path ./migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		$(action)