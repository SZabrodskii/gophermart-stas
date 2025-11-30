# Накопительная система лояльности «Гофермарт»

Дипломный проект курса «Go-разработчик» - система лояльности для интернет-магазина «Гофермарт».

## 📋 Соответствие техническому заданию

### ✅ Полностью реализовано

**HTTP API эндпоинты:**
- `POST /api/user/register` — регистрация пользователя
- `POST /api/user/login` — аутентификация пользователя  
- `POST /api/user/orders` — загрузка номера заказа для расчёта
- `GET /api/user/orders` — получение списка загруженных номеров заказов
- `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности
- `POST /api/user/balance/withdraw` — запрос на списание баллов
- `GET /api/user/withdrawals` — получение информации о выводе средств

**Бизнес-логика:**
- ✅ Регистрация, аутентификация и авторизация пользователей (JWT)
- ✅ Приём номеров заказов от зарегистрированных пользователей
- ✅ Учёт и ведение списка заказов пользователя (PostgreSQL)
- ✅ Учёт и ведение накопительного счёта пользователя
- ✅ Проверка номеров заказов через систему начислений (Worker)
- ✅ Начисление вознаграждений на счёт пользователя
- ✅ Алгоритм Луна для валидации номеров заказов
- ✅ Статусы заказов: NEW, PROCESSING, INVALID, PROCESSED

**Конфигурация:**
- ✅ `RUN_ADDRESS` / `-a` — адрес и порт сервера
- ✅ `DATABASE_URI` / `-d` — строка подключения к PostgreSQL  
- ✅ `ACCRUAL_SYSTEM_ADDRESS` / `-r` — адрес системы начислений

**HTTP коды и форматы:**
- ✅ Правильные коды ответов согласно спецификации
- ✅ JSON для API, text/plain для загрузки заказов
- ✅ Формат времени RFC3339
- ✅ Поддержка сжатия данных (gzip)

## 🏗️ Архитектура

**Технологический стек:**
- **Gin** — HTTP фреймворк (заменён httpbara для стабильности)
- **PostgreSQL + GORM** — база данных и ORM
- **Zap** — структурированное логирование
- **Fx** — dependency injection
- **JWT** — аутентификация
- **Алгоритм Луна** — валидация номеров заказов

**Архитектурные слои:**
```
cmd/gophermart/          - точка входа приложения
internal/
├── config/              - конфигурация
├── server/              - HTTP сервер (Gin)  
├── controllers/         - HTTP контроллеры
├── services/            - бизнес-логика
├── models/              - модели данных
├── database/            - работа с БД
├── auth/                - JWT аутентификация
├── utils/               - утилиты (Luhn)
├── workers/             - фоновые процессы
└── accrual/             - клиент системы начислений
```

## 🚀 Запуск

### Быстрый старт

```bash
# Сборка
go build -o cmd/gophermart/gophermart cmd/gophermart/main.go

# Запуск с параметрами по умолчанию
./cmd/gophermart/gophermart

# Или с кастомными параметрами
./cmd/gophermart/gophermart \
  -a :8080 \
  -d "postgres://user:pass@localhost/gophermart?sslmode=disable" \
  -r "http://localhost:8081"
```

### Переменные окружения

```bash
export RUN_ADDRESS=":8080"
export DATABASE_URI="postgres://user:pass@localhost/gophermart?sslmode=disable"  
export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081"
```

## 🧪 Тестирование

### Unit тесты

```bash
# Тесты алгоритма Луна
go test ./test/luhn_test.go -v

# Тесты моделей данных
go test ./test/models_test.go -v

# Тесты соответствия спецификации
go test ./test/specification_compliance_test.go -v

# Тесты конфигурации
go test ./test/config_test.go -v
```

### Smoke test

```bash
# Комплексная проверка готовности к CI
./smoke_test.sh
```

### Интеграционные тесты

```bash
# Требует настройку PostgreSQL
go test ./test/integration_test.go -v
```

## 🔧 CI/CD и автотесты

### Статус автотестов

Автотесты **актуальны** (последняя синхронизация с шаблоном):
- ✅ `.github/workflows/gophermart.yml` - тесты функциональности  
- ✅ `.github/workflows/statictest.yml` - статический анализ

### Обновление автотестов

```bash
# Получить обновления шаблона
git fetch template && git checkout template/master .github

# Проверить изменения
git diff HEAD -- .github/

# Применить при необходимости
git add .github && git commit -m "Update autotests from template"
```

### Известные проблемы

⚠️ **Версия Go**: Автотесты используют Go 1.24, локальная разработка на 1.25.2
- **Решение**: Код совместим с Go 1.24+ (указано в go.mod)
- **Статус**: Работает в CI, проблема только с локальным запуском statictest

## 📊 API Примеры

### Регистрация
```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login":"user","password":"pass"}'
```

### Загрузка заказа  
```bash
curl -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: text/plain" \
  -d "12345678903"
```

### Получение баланса
```bash
curl http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer <token>"
```

## 🛠️ Разработка

### Требования
- Go 1.24+
- PostgreSQL 12+
- Make (опционально)

### Структура проекта
```
.
├── cmd/gophermart/          # Основное приложение
├── internal/                # Внутренние пакеты
├── test/                    # Тесты
├── .github/workflows/       # CI/CD
├── smoke_test.sh           # Smoke тесты
└── SPECIFICATION.md        # Техническое задание
```

### Полезные команды

```bash
# Проверка кода
go vet ./...
go mod verify

# Форматирование
go fmt ./...

# Сборка для разных ОС
GOOS=linux go build -o gophermart cmd/gophermart/main.go
```

---

## 📈 Миграция с httpbara на Gin

Проект успешно мигрирован с кастомного фреймворка httpbara на стандартный Gin:

**Причины миграции:**
- Проблемы с JSON сериализацией в httpbara
- Стабильность и поддержка сообщества Gin
- Лучшая совместимость с экосистемой Go

**Что сохранено:**
- ✅ Вся бизнес-логика
- ✅ Структура проекта  
- ✅ API endpoints
- ✅ Dependency injection (Fx)
- ✅ Логирование (Zap)

**Что изменено:**
- HTTP фреймворк: httpbara → Gin
- Контроллеры: адаптированы под Gin Context
- Middleware: переписан для Gin
- Роутинг: использует Gin Router

## 🎯 Начало работы

1. Склонируйте репозиторий в любую подходящую директорию
2. Настройте PostgreSQL
3. Установите зависимости: `go mod download`
4. Запустите: `go run cmd/gophermart/main.go`

## 📞 Поддержка

Для обновления автотестов из шаблона:

```bash
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
git fetch template && git checkout template/master .github
```
