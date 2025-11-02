# Makefile for wallet-basic-operations

# --- Переменные для тестов ---
TEST_COMPOSE_FILE := docker-compose.test.yml
TEST_ENV := DB_HOST=localhost MIGRATIONS_PATH=../../migrations
TEST_CMD := go test -v -race -timeout 120s
TEST_UNIT := ./tests/unit/...
TEST_INTEGRATION := ./tests/integration/...
DEBUG_CMD := go run ./cmd/app/main.go
DEBUG_ENV := DB_HOST=localhost MIGRATIONS_PATH=./migrations

# .PHONY указывает, что эти цели не являются файлами
.PHONY: help test test-unit test-integration test-coverage lint fmt tidy up-db down ci run stop test-and-run

# Выполнять каждую цель в одной оболочке для корректной работы trap
.ONESHELL:

# Цель по умолчанию - показать справку
default: help

# --- Основные команды ---

test-and-run: test run ## Запустить тесты, и в случае успеха - поднять готовое приложение

debug: ## Запустить тестовую БД и приложение с тестовым конфигом
	@echo "-> Running debug cycle..."
	trap '$(MAKE) down' EXIT; \
	$(MAKE) up-db; \
	echo "--> Waiting for the database to be ready..."; \
	sleep 3; \
	echo "--> Running Go app..."; \
	$(DEBUG_ENV) $(DEBUG_CMD) ./...

test: ## Запустить полный цикл тестов (БД -> тесты -> очистка)
	@echo "-> Running full test cycle..."
	trap '$(MAKE) down' EXIT; \
	$(MAKE) up-db; \
	echo "--> Waiting for the database to be ready..."; \
	sleep 3; \
	echo "--> Running unit tests..."; \
	$(TEST_CMD) $(TEST_UNIT); \
	echo "--> Running integration tests..."; \
	$(TEST_ENV) $(TEST_CMD) $(TEST_INTEGRATION)
	@echo "✅ Tests passed successfully!"

test-unit: ## Запустить только unit тесты
	@echo "-> Running unit tests..."
	$(TEST_CMD) $(TEST_UNIT)
	@echo "✅ Unit tests passed successfully!"

test-integration: ## Запустить только integration тесты (требует БД)
	@echo "-> Running integration tests..."
	trap '$(MAKE) down' EXIT; \
	$(MAKE) up-db; \
	echo "--> Waiting for the database to be ready..."; \
	sleep 3; \
	$(TEST_ENV) $(TEST_CMD) $(TEST_INTEGRATION)
	@echo "✅ Integration tests passed successfully!"

test-coverage: ## Запустить тесты и сгенерировать отчет о покрытии
	@echo "-> Running tests with coverage..."
	trap '$(MAKE) down' EXIT; \
	$(MAKE) up-db; \
	echo "--> Waiting for the database to be ready..."; \
	sleep 3; \
	echo "--> Running unit tests with coverage..."; \
	$(TEST_CMD) -coverprofile=coverage-unit.out $(TEST_UNIT); \
	echo "--> Running integration tests..."; \
	$(TEST_ENV) $(TEST_CMD) $(TEST_INTEGRATION); \
	go tool cover -html=coverage-unit.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# --- Команды для запуска приложения ---

setup: ## Настроить окружение: установить линтер и скачать зависимости
	@echo "-> Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "-> Downloading Go modules..."
	@go mod download

run: ## Собрать и запустить приложение + БД из docker-compose.yml
	@echo "-> Starting application and database..."
	@docker-compose up --build -d

stop: ## Остановить и удалить контейнеры приложения и БД
	@echo "-> Stopping application and database..."
	@docker-compose down -v

# --- Утилиты для разработки ---

lint: ## Проверить код линтером golangci-lint
	@echo "-> Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "!! golangci-lint is not installed. Please run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run

fmt: ## Отформатировать Go код
	@echo "-> Formatting code..."
	@go fmt ./...

tidy: ## Привести в порядок зависимости в go.mod
	@echo "-> Tidying modules..."
	@go mod tidy

# --- Вспомогательные Docker-команды для тестов ---

up-db: ## [HELPER] Поднять тестовую БД в фоновом режиме
	@echo "-> Starting test database..."
	@docker-compose -f $(TEST_COMPOSE_FILE) up -d

down: ## [HELPER] Остановить и удалить контейнеры тестовой БД
	@echo "-> Stopping test database..."
	@docker-compose -f $(TEST_COMPOSE_FILE) down -v --remove-orphans

help: ## Показать это справочное сообщение
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'