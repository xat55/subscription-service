# Subscription Service

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Технологии

- Go 1.22
- PostgreSQL 15
- Gin Web Framework
- Docker & Docker Compose
- Swagger документация

## Требования

- Docker и Docker Compose
- Make (опционально)
- Go 1.22+ (для локальной разработки)

## Быстрый запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/xat55/subscription-service
cd subscription-service
```

### 2. Запуск через Docker Compose

```bash
docker-compose up --build
```

Сервер будет доступен по адресу: http://localhost:8080

## API Эндпоинты

| Метод  | Эндпоинт                           | Описание                    |
|--------|------------------------------------|-----------------------------|
| POST   | `/api/v1/subscriptions`            | Создание подписки           |
| GET    | `/api/v1/subscriptions/:id`        | Получение подписки по ID    |
| PUT    | `/api/v1/subscriptions/:id`        | Обновление подписки         |
| DELETE | `/api/v1/subscriptions/:id`        | Удаление подписки           |
| GET    | `/api/v1/subscriptions`            | Список всех подписок        |
| GET    | `/api/v1/subscriptions/total-cost` | Подсчёт стоимости за период |
| GET    | `/health`                          | Проверка здоровья сервиса   |

## Swagger UI

Интерактивная документация API доступна после запуска сервера:

```
http://localhost:8080/swagger/index.html
```

Реализована через `swaggo/gin-swagger`. Поддерживает все CRUD-эндпоинты с примерами запросов и ответов.

### Генерация документации (для разработчиков)

Swagger-спецификация хранится в `docs/` и генерируется автоматически из Go-аннотаций:

```bash
swag init -g cmd/app/main.go --output docs/
```

Требуется установленный `swag` CLI:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

После изменения аннотаций в хендлерах — перегенерировать docs и пересобрать Docker-образ:

```bash
swag init -g cmd/app/main.go --output docs/ && docker compose up --build -d
```

### Примеры запросов

### Создание подписки

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "2024-01-01T00:00:00Z"
  }'
```

Ответ:
```bash
{
  "id": "fac8024a-c58b-47e1-b32e-ea19c54de13f",
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "2024-01-01T00:00:00Z",
  "created_at": "2026-05-25T10:17:47.250752Z",
  "updated_at": "2026-05-25T10:17:47.250752Z"
}
```

### Получение всех подписок
```bash
curl http://localhost:8080/api/v1/subscriptions
```

### Подсчёт суммарной стоимости за период
```bash
curl "http://localhost:8080/api/v1/subscriptions/total-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&start_date=2024-01-01&end_date=2024-12-31"
```

Ответ:
```json
{
  "total_cost": 4800
}
```

### Обновление подписки
```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/fac8024a-c58b-47e1-b32e-ea19c54de13f \
  -H "Content-Type: application/json" \
  -d '{
    "price": 500
  }'
```

### Удаление подписки
```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/fac8024a-c58b-47e1-b32e-ea19c54de13f
```

### Миграции базы данных

Файлы миграций находятся в папке migrations:

001_create_subscriptions_table.up.sql — создание таблицы
001_create_subscriptions_table.down.sql — откат миграции

Применение миграции вручную
```bash
docker exec -i subscription-service-postgres-1 psql -U postgres -d subscriptions < migrations/001_create_subscriptions_table.up.sql
```