# Employee Management System

REST API для управления сотрудниками на Go с веб-интерфейсом, построенный по стандарту [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

## 📋 Содержание

- [Технические требования](#технические-требования)
- [Структура проекта](#структура-проекта)
- [Установка и запуск](#установка-и-запуск)
- [API документация](#api-документация)
- [Архитектура](#архитектура)
- [Мониторинг и логирование](#мониторинг-и-логирование)
- [Тестирование](#тестирование)

##  Технические требования

- **Язык**: Go 1.21+
- **Структура проекта**: [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- **Логирование**: `slog` + JSON в stdout и файл
- **Трассировка**: OpenTelemetry (Jaeger)
- **Метрики**: Prometheus
- **Тестирование**: Нагрузочное тестирование с помощью `plow`
- **Хранилище**: In-memory (можно легко заменить на БД)

##  Структура проекта

```
employee-management/
│
├── cmd/                          # Точки входа приложения
│   └── server/
│       ├── main.go              # Минимальный main - только инициализация
│       └── static/              # Статические файлы (встроены в бинарник)
│           ├── index.html       # Веб-интерфейс
│           └── style.css       # Стили
│
├── internal/                     # Приватный код приложения (не импортируется извне)
│   ├── handler/                 # HTTP слой
│   │   └── handler.go           # Обработчики HTTP запросов, роутинг (Gin)
│   │
│   ├── service/                 # Бизнес-логика
│   │   └── employee.go          # Сервис сотрудников (валидация, бизнес-правила)
│   │
│   ├── repository/              # Слой данных
│   │   ├── repository.go        # Интерфейс репозитория
│   │   └── memory.go            # In-memory реализация (можно заменить на БД)
│   │
│   ├── models/                  # Модели данных
│   │   └── models.go            # Структуры: Employee, Department, APIResponse и т.д.
│   │
│   ├── logger/                  # Логирование
│   │   └── logger.go            # Настройка логгера (slog)
│   │
│   └── telemetry/               # Мониторинг
│       ├── metrics.go           # Prometheus метрики
│       └── tracing.go           # OpenTelemetry трассировка
│
├── logs/                        # Логи приложения
│   └── app.log                 # JSON логи
│
├── metrics/                     # Метрики
│   └── metrics.log              # Дамп метрик в файл
│
├── scripts/                     # Скрипты
│   └── load-test.sh            # Скрипт нагрузочного тестирования
│
├── plow-results/                # Результаты нагрузочного тестирования
│   ├── report.md
│   └── *.log
│
├── go.mod                       # Go модули
├── go.sum                       # Checksums зависимостей
├── server.exe                  # Скомпилированный бинарник
└── README.md                    # Этот файл
```

###  Принцип разделения по слоям

```
HTTP запрос
    ↓
[Handler] ──→ Валидация запроса, парсинг JSON
    ↓
[Service] ──→ Бизнес-логика, валидация данных
    ↓
[Repository] ──→ CRUD операции с данными
    ↓
[Memory/DB] ──→ Хранилище
```

**Зависимости идут внутрь**: Handler → Service → Repository

Это позволяет:
-  Тестировать каждый слой отдельно (моки)
-  Менять реализацию без переписывания всего кода
-  Легко читать и сопровождать код

##  Установка и запуск

### Требования

- Go 1.21 или выше
- Git

### Установка

```bash
# Клонировать репозиторий
git clone <repository-url>
cd Logger-na-GO-KIP-version_2

# Установить зависимости
go mod download
```

### Сборка

```bash
# Собрать бинарник
go build -o server.exe ./cmd/server

# Или запустить напрямую
go run ./cmd/server
```

### Запуск

```bash
# Запустить сервер
./server.exe

# Сервер запустится на http://localhost:8080
```

### Проверка работоспособности

```bash
# Запустить тестовый скрипт
test_server.bat

# Или проверить вручную
curl http://localhost:8080/api/health
```

##  API документация

### Базовый URL
http://localhost:8080/api


### Эндпоинты

#### 1. Health Check

```http
GET /api/health
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-12-21T17:46:31Z",
    "service": "employee-management-system"
  }
}
```

#### 2. Получить список департаментов

```http
GET /api/departments
```

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "dept1",
      "name": "IT-департамент",
      "description": "Разработка ПО",
      "created_at": "2025-12-21T17:46:31Z"
    }
  ]
}
```

#### 3. Получить сотрудников департамента

```http
GET /api/employees/department/:departmentId
```

**Пример:**
```http
GET /api/employees/department/dept1
```

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "emp1",
      "full_name": "Иванов Иван Иванович",
      "gender": "male",
      "age": 35,
      "education": "higher",
      "position": "Программист",
      "passport": "1234 567890",
      "department_id": "dept1",
      "status": "active",
      "created_at": "2025-12-21T17:46:31Z",
      "updated_at": "2025-12-21T17:46:31Z"
    }
  ]
}
```

#### 4. Поиск сотрудников

```http
POST /api/employees/search
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "full_name": "Иванов",
  "position": "Программист",
  "gender": "male",
  "education": "higher",
  "age_from": 25,
  "age_to": 40
}
```

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "emp1",
      "full_name": "Иванов Иван Иванович",
      ...
    }
  ]
}
```

#### 5. Создать сотрудника

```http
POST /api/employees
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "full_name": "Петров Петр Петрович",
  "gender": "male",
  "age": 30,
  "education": "higher",
  "position": "Аналитик",
  "passport": "5678 901234",
  "department_id": "dept1",
  "status": "active"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "emp5",
    "full_name": "Петров Петр Петрович",
    ...
  },
  "message": "Сотрудник успешно создан"
}
```

#### 6. Обновить сотрудника

```http
PUT /api/employees/:id
Content-Type: application/json
```

**Тело запроса:** (те же поля, что и при создании)

#### 7. Изменить статус сотрудника

```http
PATCH /api/employees/:id/status
Content-Type: application/json
```

**Тело запроса:**
```json
{
  "status": "vacation"  // или "active", "fired"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Сотрудник отправлен в отпуск"
}
```

#### 8. Получить список должностей

```http
GET /api/positions
```

**Ответ:**
```json
{
  "success": true,
  "data": [
    "Программист",
    "Аналитик",
    "Тестировщик",
    ...
  ]
}
```

#### 9. Получить статистику сотрудников

```http
GET /api/metrics
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "timestamp": "2025-12-21T17:46:31Z",
    "stats": {
      "total": 4,
      "by_status": {
        "active": 3,
        "vacation": 1
      },
      "by_department": {
        "dept1": 2,
        "dept2": 1,
        "dept3": 1
      }
    },
    "message": "Метрики обновлены"
  }
}
```

### Веб-интерфейс

Откройте в браузере: `http://localhost:8080`

Веб-интерфейс позволяет:
- Просматривать сотрудников по департаментам
- Создавать новых сотрудников
- Редактировать данные сотрудников
- Изменять статус (активен, в отпуске, уволен)
- Искать сотрудников по различным фильтрам

## 🏛️ Архитектура

### Слои приложения

1. **Handler** (`internal/handler/`)
   - Обработка HTTP запросов
   - Парсинг JSON
   - Формирование ответов
   - Middleware (логирование, трассировка)

2. **Service** (`internal/service/`)
   - Бизнес-логика
   - Валидация данных
   - Проверка бизнес-правил

3. **Repository** (`internal/repository/`)
   - Абстракция доступа к данным
   - CRUD операции
   - Текущая реализация: in-memory
   - Легко заменить на PostgreSQL, MySQL и т.д.

4. **Models** (`internal/models/`)
   - Структуры данных
   - DTO (Data Transfer Objects)

### Зависимости

```
main.go
  ├── handler (HTTP слой)
  │     └── service (Бизнес-логика)
  │           └── repository (Данные)
  ├── logger (Логирование)
  └── telemetry (Метрики и трассировка)
```

##  Мониторинг и логирование

### Логирование

Логи пишутся в:
- **Консоль** (stdout) - JSON формат
- **Файл** `logs/app.log` - JSON формат

**Формат лога:**
```json
{
  "time": "2025-12-21T17:46:31.3270798+03:00",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/departments",
  "status": 200,
  "duration": "0s",
  "client_ip": "::1"
}
```

### Метрики Prometheus

Доступны по адресу: `http://localhost:8080/metrics`

**Метрики:**
- `http_requests_total` - общее количество HTTP запросов
- `http_request_duration_seconds` - длительность запросов
- `employees_total` - общее количество сотрудников
- `employees_by_status` - количество сотрудников по статусам

**Пример:**
```prometheus
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/departments",status="200"} 657716
```

### Трассировка OpenTelemetry

Поддержка Jaeger (настраивается через переменные окружения).

##  Тестирование

### Нагрузочное тестирование

Используется инструмент [plow](https://github.com/six-ddc/plow).

**Пример:**
```bash
# Тест получения департаментов
plow http://localhost:8080/api/departments -c 10 -d 30s

# Тест поиска сотрудников
plow http://localhost:8080/api/employees/search \
  -c 20 -n 1500 \
  -T "application/json" \
  --body '{"full_name":"Иванов"}' \
  -m POST
```

**Результаты тестирования:**
- При 1 соединении: ~4910 RPS
- При 5 соединениях: ~6261 RPS
- При 10 соединениях: ~5375 RPS

Подробные результаты в папке `plow-results/`.

### Запуск тестов

```bash
# Windows
run_tests_fixed.bat

# Linux/Mac
./scripts/load-test.sh
```

##  Технологии

- **Go 1.21+** - основной язык
- **Gin** - HTTP веб-фреймворк
- **slog** - структурированное логирование
- **Prometheus** - метрики
- **OpenTelemetry** - трассировка
- **Vue.js 3** - фронтенд (встроен в статические файлы)


