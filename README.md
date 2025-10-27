# Wallet Basic Operations

Сервис для управления кошельками с базовыми операциями пополнения и снятия средств.

## Описание проекта

Wallet Basic Operations - это REST API сервис, написанный на Go, который предоставляет следующие возможности:

- Создание новых кошельков
- Пополнение кошельков
- Снятие средств с кошельков
- Проверка баланса кошельков
- Проверка работоспособности сервиса

Сервис использует PostgreSQL в качестве базы данных и предоставляет API, соответствующее спецификации OpenAPI 3.0.

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
```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000"
  }'
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

## Разработка

### Структура проекта

```
├── api/                    # OpenAPI спецификация
├── cmd/                    # Точка входа приложения
├── internal/               # Внутренний код приложения
│   ├── app/               # Конфигурация и запуск сервера
│   ├── config/            # Управление конфигурацией
│   ├── generated/         # Сгенерированный код из OpenAPI
│   ├── handlers/          # Обработчики HTTP запросов
│   ├── repository/        # Работа с базой данных
│   └── service/           # Бизнес-логика
├── migrations/             # Миграции базы данных
├── sql/                   # SQL скрипты инициализации
├── test/                  # Тесты
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
# Запуск всех тестов
make test

# Запуск тестов с покрытием
make test-coverage

# Запуск тестов и затем запуск приложения
make test-and-run
```

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

