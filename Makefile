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

run-analytics:
	@go mod tidy && \
	go run cmd/analytics/main.go

env-up:
	@docker compose up -d

env-dev-up:
	@docker compose up -d redpanda redis postgres

env-cleanup: ## env: Очистить окружение проекта
	@read -p "Remove all volumes?. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down postgres port-forwarder redpanda --volumes && \
		docker volume rm url-shortener_pgdata && \
		docker volume rm url-shortener_kafkadata && \
		echo "Cleanup environment files"; \
	else \
		echo "Environment cleanup cancelled"; \
	fi

env-down:
	@docker compose down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Parameter name is required. Example: make migrate-create name=init"; \
		exit 1; \
	fi; \

	docker compose run --rm postgres-migrate \
		create \
		-ext sql \
		-dir ./migrations \
		-seq $(name)

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

kafka-topic-init:
	docker compose exec redpanda rpk topic create ${KAFKA_TOPIC} \
		--if-not-exists \
		--partitions 3 \
		--replicas 1

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Parameter action is required. Example: make migrate-action action=up 1"; \
		exit 1; \
	fi; \

	docker compose run --rm postgres-migrate \
		-path ./migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		$(action)

EASYP_VERSION := v0.16.6
PROTOC_GEN_GO_VERSION := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.6.2
PROTOC_GEN_VALIDATE := v1.3.3
MOCKGEN_VERSION := v0.6.0

.PHONY: tools
tools: ## Установить easyp, protoc-плагины и mockgen зафиксированных версий
	go install github.com/easyp-tech/easyp/cmd/easyp@$(EASYP_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE)
	go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: proto
proto: ## Сгенерировать Go-код из .proto (api/proto -> api/gen)
	easyp mod download
	easyp generate

.PHONY: proto-lint
proto-lint: ## Проверить .proto на соответствие правилам easyp.yaml
	easyp lint

.PHONY: proto-breaking
proto-breaking: ## Проверить breaking changes относительно основной ветки
	easyp breaking --against master

.PHONY: proto-check
proto-check: proto ## CI-проверка: сгенерённый код не разошёлся с .proto
	@if [ -n "$$(git status --porcelain api/gen)" ]; then \
		echo "api/gen out of sync with .proto files — run 'make proto' and commit" && \
		git status --porcelain api/gen && \
		exit 1; \
	fi

gen: ## Сгенерировать все моки в проекте
	@go generate ./...

test: ## Запустить быстрые юнит-тесты
	@go test -v -short ./...