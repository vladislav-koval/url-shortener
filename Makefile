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

env-cleanup: ## env: Очистить окружение проекта
	@read -p "Remove postgres volume?. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down postgres port-forwarder kafka --volumes && \
		docker volume rm url-shortener_pgdata && \
		docker volume rm url-shortener_kafkadata && \
		echo "Cleanup environment files"; \
	else \
		echo "Environment cleanup cancelled"; \
	fi

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

kafka-topic-init:
	docker-compose exec kafka /opt/kafka/bin/kafka-topics.sh \
		--bootstrap-server localhost:9092 \
		--create --if-not-exists \
		--topic ${KAFKA_TOPIC} \
		--partitions 3 \
		--replication-factor 1

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