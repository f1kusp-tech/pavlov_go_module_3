# URL shortener (REST API)

## Быстрый старт

```bash
git clone https://github.com/f1kusp-tech/pavlov_go_module_3.git
cd pavlov_go_module_3
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

Скопируйте `.env.example` в `.env` (**обязательно** — без `.env` не подставятся переменные в `docker-compose` для Postgres).

| Параметр | Где задано | Значение |
|----------|------------|----------|
| HTTP | `config/local.yaml` + порт compose | `0.0.0.0:8082` / снаружи `localhost:8082` |
| Postgres host/port | `config/local.yaml` + `.env` | `postgres` / `5432` (`DB_HOST`, `DB_PORT`) |
| Postgres user / password / DB | **только `.env`** | `shortener` / `shortener` / `shortener` |
| Basic Auth | **только `.env`** | `myuser` / `mypassword` |

Postgres в `docker-compose` подставляет `${DB_USER}`, `${DB_PASSWORD}`, `${DB_NAME}` из `.env`. Без Docker укажите `DB_HOST=localhost` в `.env`.

## Тесты

Юнит-тесты handlers:

```bash
go test ./internal/http-server/handlers/...
```

Интеграционные тесты (сервер на `localhost:8082` должен быть запущен):

```bash
go test ./test/...
```