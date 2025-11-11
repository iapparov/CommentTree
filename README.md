# DelayedNotifier

## Описание

CommentTree — древовидные комментарии с навигацией и поиском

## Состав репозитория

- **cmd/main.go** — точка входа через FX DI.
- **internal/**
  - **app/domain** — модели данных Comment, CommentNode.
  - **app** — модели данных CommentService.
  - **config/** — загрузка конфигурации из YAML.
  - **di/** — реализация зависимостей через UberFX.
  - **storage/db** — работа с PostgreSQL (CRUD).
  - **web/** — HTTP-обработчики и роутер.
- **config/local.yaml** — пример конфигурации.
- **migrations/** — SQL-миграции для PostgreSQL.
- **docs/** — Swagger-документация.
- **web/index.html** — простая страница работы с комментариям
- **docker-compose.yml** — запуск PostgreSQL через Docker.
- **.env.example** - пример env файла для кредов.


---

## Быстрый старт

### 1. Запуск инфраструктуры

```sh
docker-compose up -d
```
(Запустит контейнеры: postgres → порт 5433)

### 2. Настроить переменные окружения и конфигурацию
(пример в .env.example + config/local.yaml)

### 3. Применить миграции (migrate):

```sh
migrate -path migrations -database "postgres://user:password@localhost:5433/dbname?sslmode=disable" up
```

### 4. Запуск сервиса

```sh
go run ./cmd/main.go
```

Сервис стартует на порту 8080.



## API

- **POST /comments** — создание комментария (с указанием родительского) JSON: parent_id, text;
- **GET /comments?parent={id}** — получение комментария и всех вложенных;
- **DELETE /comments/{id}** —  удаление комментария и всех вложенных под ним.
- **Swagger**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## Веб-интерфейс
Откройте index.html в браузере — простая страница для просмотра уведомлений/отправки тестов через API.


## Тесты
Юнит-тесты: `go test ./internal/...`

## Миграции

- `migrations/000001_create_tables.up.sql` — создание таблиц.
- `migrations/000001_create_tables.down.sql` — удаление таблиц.

---

## Логирование и метрики
Логирование реализовано через wbf/zlog (используется в internal/*).

## Зависимости

- Go 1.25+
- PostgreSQL 16+
- Docker (для локального запуска инфраструктуры)

---

## Swagger

- Swagger: [docs/swagger.yaml](docs/swagger.yaml)
- Документация генерируется автоматически и доступна по `/swagger/*`.