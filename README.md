# Wallet Basic Operations

Сервис для управления кошельками с базовыми операциями пополнения и снятия средств.

## Описание проекта

Wallet Basic Operations - это REST API сервис, написанный на Go, который предоставляет следующие возможности:

- Создание новых кошельков (UUID генерируется автоматически)
- Пополнение кошельков (Deposit)
- Снятие средств с кошельков (Withdraw)
- Проверка баланса кошельков
- Проверка работоспособности сервиса

Сервис использует PostgreSQL в качестве базы данных и предоставляет API, соответствующее спецификации OpenAPI 3.0.

## Архитектура

Проект следует принципам чистой архитектуры:

- **Handlers** (`internal/handlers/`) - обработка HTTP запросов и валидация
- **Service** (`internal/service/`) - бизнес-логика приложения
- **Repository** (`internal/repository/`) - работа с базой данных
- **Errors** (`internal/errors/`) - кастомные типизированные ошибки

## Технологический стек

- **Язык программирования**: Go 1.25
- **Веб-фреймворк**: Chi
- **База данных**: PostgreSQL
- **ORM**: pgx
- **Миграции**: Goose
- **API документация**: OpenAPI 3.0
- **Конфигурация**: переменные окружения
- **Контейнеризация**: Docker & Docker Compose
- **Тестирование**: testify

## Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)

## Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/devopesik/wallet-basic-operations
cd wallet-basic-operations
```

### 2. Настройка конфигурации

Отредактируйте `config.env` при необходимости:

### 3. Запуск приложения

Запуск тестов и приложения

Тесты выполняются локально, надо установить go 1.25  
и выполнить команды ниже
```bash
make setup # для установки зависимостей
make test-and-run # запустить тесты и приложение
```

Или же без тестов

```bash
# Запуск приложения и базы данных
make run
```

```bash
# Или с помощью Docker Compose
docker-compose up --build -d
```

Приложение будет доступно по адресу: `http://localhost:8080`

## API документация

### Эндпоинты

#### Health Check
- **GET** `/health` - Проверка работоспособности сервиса

#### Управление кошельками
- **POST** `/api/v1/wallets` - Создание нового кошелька
- **GET** `/api/v1/wallets/{walletId}` - Получение баланса кошелька
- **POST** `/api/v1/wallet` - Выполнение операции (пополнение/снятие)

### Примеры запросов

#### Создание кошелька

UUID генерируется автоматически на сервере и возвращается в ответе.

```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Ответ:**
```json
{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "balance": 0
}
```

#### Получение баланса
```bash
curl -X GET http://localhost:8080/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000
```

#### Пополнение кошелька
```bash
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000",
    "operationType": "DEPOSIT",
    "amount": 1000
  }'
```

#### Снятие средств
```bash
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000",
    "operationType": "WITHDRAW",
    "amount": 500
  }'
```

### Коды ответов и ошибки

Сервис использует стандартные HTTP коды ответов:

- **200 OK** - Успешное получение данных
- **201 Created** - Успешное создание кошелька
- **204 No Content** - Успешная операция без возврата данных
- **400 Bad Request** - Некорректный запрос (невалидный JSON, UUID, сумма, тип операции)
- **404 Not Found** - Кошелёк не найден
- **409 Conflict** - Конфликт (кошелёк уже существует, недостаточно средств)
- **500 Internal Server Error** - Внутренняя ошибка сервера

**Формат ошибки:**
```json
{
  "message": "описание ошибки"
}
```

## Разработка

### Структура проекта

```
├── api/                    # OpenAPI спецификация
├── cmd/                    # Точка входа приложения
├── internal/               # Внутренний код приложения
│   ├── app/               # Конфигурация и запуск сервера
│   ├── config/            # Управление конфигурацией
│   ├── errors/            # Кастомные типизированные ошибки
│   ├── generated/         # Сгенерированный код из OpenAPI
│   ├── handlers/          # Обработчики HTTP запросов
│   ├── repository/        # Работа с базой данных
│   │   └── postgres/      # Реализация для PostgreSQL
│   └── service/           # Бизнес-логика
├── migrations/             # Миграции базы данных
├── sql/                   # SQL скрипты инициализации
├── tests/                 # Все тесты
│   ├── unit/              # Unit тесты (без БД)
│   └── integration/      # Интеграционные тесты (с БД)
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── docker-compose.test.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Запуск тестов

```bash
# Запуск всех тестов (unit + integration)
make test

# Запуск только unit тестов (быстро, без БД)
make test-unit

# Запуск только integration тестов (требует БД)
make test-integration

# Запуск тестов с покрытием
make test-coverage

# Запуск тестов и затем запуск приложения
make test-and-run
```

**Типы тестов:**
- **Unit тесты** (`tests/unit/`) - тестирование бизнес-логики с использованием моков, не требуют БД
- **Integration тесты** (`tests/integration/`) - тестирование API с реальной БД, включая нагрузочные тесты

### Линтинг и форматирование

```bash
# Проверка кода линтером
make lint

# Форматирование кода
make fmt

# Обновление зависимостей
make tidy
```

### Генерация кода из OpenAPI

Код генерируется автоматически при сборке Docker-образа. Для локальной генерации:

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -package generated -generate types,chi-server -o internal/generated/api.gen.go ./openapi.yaml
```

## Архитектурные решения

### Кастомные ошибки

Проект использует систему кастомных типизированных ошибок (`internal/errors/`):

- Типобезопасная обработка ошибок (вместо сравнения строк)
- Каждая ошибка имеет свой HTTP статус код
- Поддержка цепочек ошибок через `Unwrap()`
- Централизованная обработка в handlers

### Структура данных

Вместо работы с отдельными полями используется структура `Wallet`:

```go
type Wallet struct {
    ID      uuid.UUID
    Balance int64
}
```

### Разделение операций

Логика пополнения и списания разделена на уровне service и repository:

- `Deposit(ctx, walletID, amount)` - пополнение баланса
- `Withdraw(ctx, walletID, amount)` - списание с проверкой достаточности средств

## База данных

### Миграции

Проект использует Goose для управления миграциями. Миграции находятся в директории `migrations/`.

Текущая схема данных:

```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);
```

### Подключение к базе данных

```bash
# Подключение к контейнеру с базой данных
docker exec -it postgres psql -U wallet_user -d wallet_db
```

## Переменные окружения

| Переменная        | Описание                        | Значение по умолчанию |
|-------------------|---------------------------------|-----------------------|
| `SERVER_PORT`     | Порт сервера                    | `8080`                |
| `DB_HOST`         | Хост базы данных                | `db`                  |
| `DB_PORT`         | Порт базы данных                | `5432`                |
| `DB_USER`         | Пользователь БД                 | `wallet_user`         |
| `DB_PASSWORD`     | Пароль пользователя БД          | `wallet_password`     |
| `DB_NAME`         | Имя базы данных                 | `wallet_db`           |
| `MIGRATIONS_PATH` | Путь до директории с миграциями | `migrations`          |

## Доступные команды Makefile

Выполните `make help` для просмотра всех доступных команд.

**Основные команды:**
```bash
make help                # Показать список всех доступных команд
make setup              # Настроить окружение (установить линтер, зависимости)
make test               # Запустить все тесты (unit + integration)
make test-unit          # Запустить только unit тесты (быстро, без БД)
make test-integration   # Запустить только integration тесты (требует БД)
make test-coverage      # Запустить тесты с генерацией отчета о покрытии
make test-and-run       # Запустить тесты и затем приложение
make run                # Запустить приложение и БД
make stop               # Остановить приложение и БД
make debug              # Запустить тестовую БД и приложение для отладки
make lint               # Проверить код линтером
make fmt                # Отформатировать код
make tidy               # Обновить зависимости
make up-db              # Поднять тестовую БД
make down               # Остановить тестовую БД
```

**Рекомендации:**
- Для быстрой проверки кода во время разработки используйте `make test-unit`
- Для полной проверки перед коммитом используйте `make test`
- Для проверки покрытия кода используйте `make test-coverage`

