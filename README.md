# Накопительная система лояльности «Гофермарт»

## Описание

Система представляет собой HTTP API для управления баллами лояльности интернет-магазина "Гофермарт". Реализована с использованием принципов Domain-Driven Design (DDD) на языке Go.

## Архитектура

Проект построен с использованием принципов чистой архитектуры и DDD:

### Слои архитектуры:

- **Domain** (`internal/domain/`) - доменные модели и интерфейсы
  - `user/` - доменная модель пользователя
  - `order/` - доменная модель заказа с алгоритмом Луна
  - `withdrawal/` - доменная модель списания

- **Service** (`internal/service/`) - бизнес-логика приложения
  - `UserService` - управление пользователями
  - `OrderService` - управление заказами  
  - `WithdrawalService` - управление списаниями
  - `AccrualProcessorService` - интеграция с внешней системой

- **Infrastructure** (`internal/infrastructure/`) - внешние интеграции
  - `http/handler/` - HTTP хендлеры
  - `repository/postgres/` - репозитории для PostgreSQL  
  - `jwt/` - JWT авторизация
  - `password/` - хеширование паролей bcrypt
  - `accrual/` - клиент внешней системы

## Технологический стек

- **Go 1.24** - основной язык
- **Chi v5** - HTTP роутер
- **SQLX** - работа с базой данных
- **PostgreSQL** - СУБД
- **JWT** - авторизация
- **bcrypt** - хеширование паролей
- **Zap** - структурированное логирование

## API Endpoints

### Публичные эндпоинты:
- `POST /api/user/register` - регистрация пользователя
- `POST /api/user/login` - аутентификация пользователя

### Приватные эндпоинты (требуют JWT токен):
- `POST /api/user/orders` - загрузка номера заказа
- `GET /api/user/orders` - получение списка заказов
- `GET /api/user/balance` - получение баланса
- `POST /api/user/balance/withdraw` - списание баллов
- `GET /api/user/withdrawals` - история списаний

## Настройка и запуск

### Переменные окружения

```bash
# Обязательные
DATABASE_URI="postgres://user:password@localhost/gophermart?sslmode=disable"

# Опциональные  
RUN_ADDRESS="localhost:8080"              # адрес и порт сервера
ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081"  # адрес системы расчета баллов
```

### Альтернативно через флаги командной строки

```bash
./gophermart -a localhost:8080 -d "postgres://..." -r "http://localhost:8081"
```

### Подготовка базы данных

1. Создайте базу данных PostgreSQL
2. Примените миграции из папки `migrations/`

```sql
-- Пример подключения к PostgreSQL и создания БД
CREATE DATABASE gophermart;
\c gophermart;
\i migrations/001_init.sql;
```

### Сборка и запуск

```bash
# Установка зависимостей
go mod tidy

# Сборка
go build -o gophermart ./cmd/gophermart

# Запуск
DATABASE_URI="postgres://user:password@localhost/gophermart?sslmode=disable" ./gophermart
```

## Использование API

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login":"user@example.com","password":"password123"}'
```

### Аутентификация

```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login":"user@example.com","password":"password123"}'
```

### Загрузка заказа (с JWT токеном)

```bash
curl -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: text/plain" \
  -d "12345678903"
```

### Получение баланса

```bash
curl -X GET http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Особенности реализации

### Валидация номеров заказов
Реализован алгоритм Луна для проверки корректности номеров заказов.

### Авторизация
Используется JWT токены в заголовке `Authorization: Bearer <token>`.

### Интеграция с внешней системой
Сервис автоматически опрашивает внешнюю систему расчета баллов каждые 10 секунд для обновления статусов заказов.

### Graceful Shutdown
Приложение корректно завершает работу при получении сигналов SIGINT/SIGTERM.

### Middleware
- Сжатие GZIP
- Логирование запросов  
- Восстановление после паник
- Таймауты запросов

## Структура базы данных

### Таблица users
- `id` UUID - идентификатор пользователя
- `login` TEXT - уникальный логин
- `password_hash` TEXT - хеш пароля
- `current_balance` NUMERIC - текущий баланс
- `withdrawn_balance` NUMERIC - сумма всех списаний

### Таблица orders  
- `id` BIGSERIAL - идентификатор заказа
- `user_id` UUID - ссылка на пользователя
- `number` TEXT - номер заказа (уникальный)
- `status` ENUM - статус обработки
- `accrual` NUMERIC - начисленные баллы

### Таблица withdrawals
- `id` BIGSERIAL - идентификатор списания  
- `user_id` UUID - ссылка на пользователя
- `order_number` TEXT - номер заказа для оплаты
- `amount` NUMERIC - сумма списания

## Разработка

### Запуск тестов
```bash
go test ./...
```

### Линтер
```bash
golangci-lint run
```
