# URL shortener (REST API)

## Быстрый старт

```bash
git clone <repository-url>
cd rest_api_go
cp .env.example .env
docker compose up
```

Сервис доступен на `http://localhost:8082`.

## Аутентификация

Защищённые маршруты (`POST/PUT/DELETE /url`) требуют Basic Auth. Логин и пароль задаются **только** в `.env` (`AUTH_USER`, `AUTH_PASSWORD`) — см. `.env.example`.

## API

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/url` | да | Создать короткую ссылку (201) |
| PUT | `/url/{alias}` | да | Обновить URL по alias |
| DELETE | `/url/{alias}` | да | Удалить по alias |
| GET | `/{alias}` | нет | Редирект на исходный URL (302) |

## Конфигурация

Переменные окружения — в `.env.example`. Для локального запуска без Docker укажите `DB_HOST=localhost` в `.env` и поднимите Postgres с теми же кредами, что в `docker-compose.yml` и `config/local.yaml`.

## Тесты

Юнит-тесты handlers:

```bash
go test ./internal/http-server/handlers/...
```

Интеграционные тесты (сервер на `localhost:8082` должен быть запущен):

```bash
go test ./test/...
```