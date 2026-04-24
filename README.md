# Modern Social Network Back

Бэкенд для социальной сети на Go с REST API, WebSocket-чатом, JWT-аутентификацией, email-верификацией и PostgreSQL.

## Что это за проект

Сервис предоставляет API для:

- регистрации, логина, refresh-токенов, email verification и email 2FA;
- профиля пользователя и админ-операций;
- постов, комментариев, лайков;
- подписок (follow/follower);
- stories;
- чата (REST + WebSocket) и presence;
- уведомлений.

## Что мы используем

### Язык и платформа

- Go 1.25

### Основной стек

- Gin (`github.com/gin-gonic/gin`) - HTTP роутинг и middleware
- GORM (`gorm.io/gorm`) + PostgreSQL драйвер (`gorm.io/driver/postgres`) - ORM и БД
- JWT (`github.com/golang-jwt/jwt/v5`) - access/refresh токены
- WebSocket (`github.com/gorilla/websocket`) - realtime чат и presence
- Swagger (`github.com/swaggo/*`) - документация API
- Dotenv (`github.com/joho/godotenv`) - загрузка `.env`
- CUID (`github.com/lucsky/cuid`) - генерация ID (в генераторе stories)
- Argon2 (`golang.org/x/crypto`) - хеширование паролей

### Архитектура

- `cmd/` - входные точки приложения и утилиты
- `internal/handlers/` - HTTP/WS handlers
- `internal/routes/` - маршрутизация по модулям
- `internal/services/` - бизнес-логика
- `internal/repository/` - доступ к данным (GORM)
- `internal/models/` - модели БД
- `internal/middleware/` - auth/admin middleware

## Быстрый старт

### 1. Требования

- Go 1.25+
- PostgreSQL 13+

### 2. Установить зависимости

```bash
go mod download
```

### 3. Настроить окружение

Проект читает `.env` автоматически при старте. Если файла нет, создайте его в корне проекта.

Пример `.env`:

```env
# Server
PORT=8080

# Auth
JWT_SECRET=change_me
ADMIN_TOKEN=change_me_admin_token
EMAIL_2FA_ENABLED=true

# Postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=social_network
DB_SSLMODE=disable

# SMTP
SMTP_HOST=localhost
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@example.com
```

### 4. Запустить сервер

```bash
go run ./cmd
```

Сервер поднимется на `http://localhost:${PORT}` (по умолчанию `http://localhost:8080`).

При старте автоматически выполняются:

- подключение к PostgreSQL;
- миграции (`AutoMigrate` + индексы + ограничения);
- периодическая очистка просроченных stories.

## Запуск через Docker

### 1. Подготовить .env

Скопируйте `.env.example` в `.env` и при необходимости измените значения:

- Linux/macOS: `cp .env.example .env`
- Windows PowerShell: `Copy-Item .env.example .env`

### 2. Поднять контейнеры

```bash
docker compose up --build -d
```

Состав:

- `app` - Go API (порт `8080`)
- `db` - PostgreSQL 16 (порт `5432`)
- `seed` - одноразовый сервис сидеров (профиль `tools`)

### 3. Проверка

- API: `http://localhost:8080/openapi.json`
- Swagger UI: `http://localhost:8080/swagger/index.html`

### 4. Остановить

```bash
docker compose down
```

Если нужно удалить и тома (данные БД и uploads):

```bash
docker compose down -v
```

### 5. Запуск сидеров через Docker

Запустить все сидеры (users/follows + stories):

```bash
docker compose --profile tools run --rm seed
```

Команда использует сервис `seed` из `docker-compose.yml` и подключается к той же БД (`db`).

## Документация API

- Swagger UI: `GET /swagger/index.html`
- OpenAPI JSON: `GET /openapi.json`
- Док по чату: `docs/chat-api.md`

Базовый префикс API: `/api/v1`

## Ключевые модули API

- `auth`:
  - `/auth/register`
  - `/auth/login`
  - `/auth/refresh`
  - `/auth/verify-email`
  - `/auth/resend-verify-email`
  - `/auth/2fa/verify`, `/auth/2fa/request`, `/auth/toggle-2fa`
- `user`: `/user/*`
- `post`: `/post/*`
- `comment`: `/comment/*`
- `follow`: `/follow/*`, `/user/:id/followers`, `/user/:id/following`
- `story`: `/story/*`
- `skill`: `/skill`
- `notifications`: `/notifications/*`
- `chat` REST: `/chat/*`
- `chat` WS: `/ws?token=<access_token>`

## Аутентификация

- Для защищенных REST-эндпоинтов: `Authorization: Bearer <access_token>`
- Refresh токен хранится в `HttpOnly` cookie `refreshToken`
- Админ-эндпоинты пользователей защищены заголовком `X-Admin-Token`

## CORS

В текущей конфигурации разрешен origin:

- `http://localhost:3000`

Если фронтенд работает на другом origin, обновите CORS-конфиг в `cmd/routes.go`.

## Сидеры и утилиты

- Заполнить пользователей и подписки:

```bash
go run ./cmd/seeder
```

- Сгенерировать stories:

```bash
go run ./cmd/gen_stories
```

- Запустить все сидеры:

```bash
go run ./cmd/seed_all
```

- Быстрая диагностика таблицы подписок:

```bash
go run ./cmd/debug_follows
```

## Статика и загрузки

- Статические файлы отдаются из `/uploads`
- Папки загрузок:
  - `uploads/avatars/`
  - `uploads/avatars/random/`

## Структура проекта (кратко)

```text
cmd/
  main.go           # bootstrap приложения
  routes.go         # регистрация маршрутов
  server.go         # HTTP сервер
  seeder/           # генерация пользователей/подписок/аватаров
  gen_stories/      # генерация stories
  seed_all/         # запуск нескольких сидеров
  debug_follows/    # отладка подписок

internal/
  handlers/         # transport слой (HTTP/WS)
  routes/           # wiring endpoints
  services/         # business logic
  repository/       # data access
  models/           # DB models
  middleware/       # auth/admin middleware
```

## Полезно знать

- Проект уже содержит `openapi.json`, `docs/swagger.json`, `docs/swagger.yaml`.
- Для продакшена обязательно задайте сильные значения `JWT_SECRET` и `ADMIN_TOKEN`.
- Убедитесь, что SMTP-провайдер настроен, иначе email verification/2FA не будут работать.
