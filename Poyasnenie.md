# Анализ проекта Logger-na-GO-KIP и руководство по внедрению библиотеки `logger_kim`

---

## 1. Общий обзор проекта

### Структура проекта

```
Logger-na-GO-KIP/
├── cmd/server/main.go          # Точка входа, запуск HTTP-сервера
├── internal/
│   ├── handler/handler.go      # HTTP-обработчики (Gin), middleware логирования
│   ├── logger/logger.go        # Текущая инициализация логгера (slog stdlib)
│   ├── service/employee_service.go  # Бизнес-логика, активно использует slog
│   ├── models/                 # Модели данных
│   ├── repository/             # Слой доступа к данным (in-memory)
│   └── telemetry/metrics.go    # Prometheus метрики, запись в файл
├── logger_kim/                 # Библиотека логирования (цель внедрения)
│   ├── logger.go               # Интерфейс Logger и реализация kimLogger
│   ├── config.go               # Config, Level, Format
│   ├── features.go             # Декларативные Feature-флаги
│   ├── handler.go              # Интерфейс Handler
│   ├── stdout_handler.go       # StdoutHandler, StderrHandler
│   ├── file_handler.go         # FileHandler с ротацией по размеру
│   ├── async_handler.go        # AsyncHandler — асинхронная буферизация
│   ├── sampling.go             # SamplingHandler — ограничение потока логов
│   ├── slog_bridge.go          # SlogHandler — мост к log/slog
│   ├── opensearch_handler.go   # OpenSearchHandler — Bulk API
│   ├── formatter.go            # JSON, Text, Pretty, Logfmt форматтеры
│   └── go.mod                  # Отдельный Go-модуль
├── go.mod                      # Модуль основного проекта: employee-management
├── docker-compose.yml
└── Makefile
```

### Технологический стек

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.21 |
| HTTP-фреймворк | Gin v1.9.1 |
| Метрики | Prometheus + OpenTelemetry |
| Текущий логгер | `log/slog` (стандартная библиотека) |
| Репозиторий | In-memory (без базы данных) |

---

## 2. Текущее состояние логирования

### Где и как используется `slog` прямо сейчас

#### `internal/logger/logger.go` — точка инициализации

```go
// Текущая реализация: ручная настройка slog через io.MultiWriter
func Setup(logDir, logFile string) (*os.File, error) {
    // Создаёт файл + пишет одновременно в stdout и файл
    multiWriter := io.MultiWriter(os.Stdout, file)
    logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)  // Устанавливает глобальный slog-логгер
    return file, nil
}
```

**Проблемы текущей реализации:**
- Нет ротации файлов — при долгой работе файл вырастет бесконечно
- Нет цветного вывода для разработки
- Уровень логирования зафиксирован (`LevelInfo`), не меняется без перезапуска
- Нет асинхронной записи — каждый лог блокирует поток

#### `internal/handler/handler.go` — HTTP middleware

```go
// loggingMiddleware использует глобальный slog
slog.Info("HTTP request",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "status", c.Writer.Status(),
    "duration", duration.String(),
    "client_ip", c.ClientIP(),
)

// handleError использует slog.Error
slog.Error("API error", "error", err, "status", status)
```

#### `internal/service/employee_service.go` — бизнес-логика

```go
// Везде slog.DebugContext(ctx, ...) с контекстными полями
slog.DebugContext(ctx, "creating employee", "employee", emp.FullName)
slog.DebugContext(ctx, "updating employee status", "employee_id", id, "status", status)
```

#### `internal/telemetry/metrics.go`

```go
slog.Error("Ошибка сбора метрик", "error", err)
```

**Итого:** `slog` используется в 4 файлах, ~15 вызовов. Все вызовы идут через **глобальный `slog`** (через `slog.SetDefault`).

---

## 3. Анализ библиотеки `logger_kim`

### Что такое `logger_kim`

`logger_kim` — кастомная библиотека структурированного логирования для Go, построенная **поверх** `log/slog`. Она не заменяет `slog`, а расширяет его через plug-in архитектуру.

**Модуль:** `github.com/metal-telo/Logger-na-GO-KIP/logger_kim`

### Ключевые компоненты

#### Интерфейс `Logger`

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    Fatal(msg string, args ...any)
    With(args ...any) Logger          // дочерний логгер с полями
    WithContext(ctx context.Context) Logger
    SetLevel(level Level)
    GetLevel() Level
    Close() error
}
```

#### Интерфейс `Handler`

```go
type Handler interface {
    Handle(ctx context.Context, entry Entry) error
    Close() error
    Enabled(level Level) bool
    SetLevel(level Level)
}
```

#### Доступные Handler-ы

| Handler | Описание |
|---------|----------|
| `StdoutHandler` | Вывод в os.Stdout, поддерживает форматы JSON/Text/Pretty/Logfmt |
| `StderrHandler` | Вывод в os.Stderr |
| `FileHandler` | Запись в файл с ротацией по размеру (MaxSize MB, MaxBackups копий) |
| `AsyncHandler` | Обёртка над любым Handler — асинхронная буферизация через канал |
| `SamplingHandler` | Обёртка — семплирование (пропускает N за период, остальные отбрасывает) |
| `OpenSearchHandler` | Отправка в OpenSearch через Bulk API |

#### Форматы вывода

| Format | Описание |
|--------|----------|
| `FormatJSON` | Стандартный JSON, совместимый с ELK/OpenSearch |
| `FormatText` | Человекочитаемый текст |
| `FormatPretty` | Цветной вывод для терминала (через `github.com/fatih/color`) |
| `FormatLogfmt` | key=value формат |

#### Feature-флаги (`Features`)

```go
type Features struct {
    Async            bool   // AsyncHandler для всех outputs
    Sampling         bool   // SamplingHandler для всех outputs
    SlogBridge       bool   // Автоматически вызывает SetAsDefault(logger)
    PrettyColor      bool   // ANSI-цвета в Pretty-форматтере
    StructuredFields bool   // Структурированные поля key=value
    MultipleOutputs  bool   // Разрешает несколько Handler-ов
    FileOutput       bool   // Разрешает FileHandler-ы
    OpenSearchOutput bool   // Разрешает OpenSearchHandler
}
```

#### Критически важный `SlogBridge`

`SlogBridge` — это **ключевой механизм бесшовной интеграции**. При `Features.SlogBridge = true`, внутри `New()` автоматически вызывается `SetAsDefault(logger)`, который устанавливает `logger_kim` как бэкенд для глобального `slog`. После этого **весь существующий код с `slog.Info`, `slog.Debug` и т.д. начинает работать через `logger_kim` без каких-либо изменений**.

```go
// slog_bridge.go — SlogHandler реализует slog.Handler поверх logger_kim.Logger
type SlogHandler struct {
    logger Logger
    attrs  []slog.Attr
    group  string
}
// Конвертирует slog.Record → logger_kim.Entry и делегирует в logger_kim
```

---

## 4. Пошаговое руководство по внедрению

### Шаг 1: Подключение `logger_kim` как локальной зависимости

Поскольку `logger_kim` находится в той же репозитории, нужно добавить `replace`-директиву в `go.mod` основного проекта.

**Файл: `go.mod`** — добавить в конец:

```go
require (
    // ... существующие зависимости ...
    github.com/metal-telo/Logger-na-GO-KIP/logger_kim v0.0.0
    github.com/fatih/color v1.19.0
)

replace github.com/metal-telo/Logger-na-GO-KIP/logger_kim => ./logger_kim
```

Затем выполнить:

```bash
go mod tidy
```

### Шаг 2: Замена `internal/logger/logger.go`

Это **единственный файл**, который нужно изменить для базовой интеграции. Весь остальной код проекта продолжает использовать `slog.*` без изменений благодаря `SlogBridge`.

**Текущий код:**

```go
package logger

import (
    "fmt"
    "io"
    "log/slog"
    "os"
)

func Setup(logDir, logFile string) (*os.File, error) {
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return nil, fmt.Errorf("не удалось создать директорию логов: %w", err)
    }
    logFilePath := logDir + "/" + logFile
    file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, fmt.Errorf("не удалось открыть файл логов: %w", err)
    }
    multiWriter := io.MultiWriter(os.Stdout, file)
    logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)
    return file, nil
}
```

**Новый код с `logger_kim`:**

```go
package logger

import (
    "fmt"

    lk "github.com/metal-telo/Logger-na-GO-KIP/logger_kim"
)

// KimLogger — хранит ссылку для вызова Close() при завершении
var KimLogger lk.Logger

// Setup инициализирует logger_kim с двумя output-ами: stdout (Pretty) и файл (JSON).
// SlogBridge автоматически подменяет глобальный slog — весь существующий код
// (slog.Info, slog.Debug, slog.Error) продолжает работать без изменений.
func Setup(logDir, logFile string) (func(), error) {
    logFilePath := logDir + "/" + logFile

    // Handler 1: цветной Pretty-вывод в терминал (для разработки)
    stdoutHandler := lk.NewStdoutHandlerWithFormat(lk.FormatPretty)

    // Handler 2: JSON-файл с ротацией (для production)
    fileHandler, err := lk.NewFileHandlerWithConfig(lk.FileHandlerConfig{
        FilePath:   logFilePath,
        MaxSize:    100,    // 100 MB — после этого создаётся новый файл
        MaxBackups: 5,      // хранить 5 старых файлов
        Format:     lk.FormatJSON,
        Level:      lk.LevelInfo,
    })
    if err != nil {
        return nil, fmt.Errorf("не удалось создать FileHandler: %w", err)
    }

    cfg := lk.Config{
        Level:       lk.LevelDebug,
        ServiceName: "employee-management",
        Outputs:     []lk.Handler{stdoutHandler, fileHandler},
        Features: lk.Features{
            SlogBridge:       true,  // ← ключевое: подменяет глобальный slog
            PrettyColor:      true,
            StructuredFields:  true,
            MultipleOutputs:  true,
            FileOutput:       true,
        },
    }

    KimLogger = lk.New(cfg)

    // Возвращаем функцию cleanup вместо *os.File
    return func() { _ = KimLogger.Close() }, nil
}
```

### Шаг 3: Обновление `cmd/server/main.go`

Изменения минимальны — только место вызова `Setup` и обработка нового возвращаемого значения:

**Текущий код:**

```go
logFile, err := logger.Setup(LogDir, LogFile)
if err != nil {
    return fmt.Errorf("ошибка настройки логгера: %w", err)
}
defer logFile.Close()
```

**Новый код:**

```go
closeLogger, err := logger.Setup(LogDir, LogFile)
if err != nil {
    return fmt.Errorf("ошибка настройки логгера: %w", err)
}
defer closeLogger()
```

Все остальные вызовы `slog.Info(...)`, `slog.Error(...)`, `slog.Debug(...)` в проекте **остаются без изменений** — они автоматически перенаправляются в `logger_kim` через `SlogBridge`.

---

## 5. Опциональные улучшения (для production)

### 5.1 Асинхронная запись в файл

При высокой нагрузке запись в файл может блокировать обработку запросов. `AsyncHandler` решает это:

```go
fileHandler, _ := lk.NewFileHandlerWithConfig(...)

// Оборачиваем fileHandler в AsyncHandler
asyncFileHandler := lk.NewAsyncHandler(fileHandler, lk.AsyncOptions{
    BufferSize:   4096,           // буфер на 4096 записей
    FlushTimeout: 5 * time.Second,
    DropOnFull:   false,          // не терять логи при перегрузке
})

cfg := lk.Config{
    Outputs: []lk.Handler{stdoutHandler, asyncFileHandler},
    // ...
}
```

> **Важно:** при использовании AsyncHandler обязательно вызывать `defer KimLogger.Close()` — иначе буфер не будет сброшен при завершении программы.

### 5.2 Семплирование для высоконагруженных эндпоинтов

Если GET-запросы идут тысячами в секунду, `SamplingHandler` сокращает объём логов:

```go
sampledStdout := lk.NewSamplingHandler(stdoutHandler, lk.SamplingConfig{
    Tick:       time.Second,
    First:      20,   // первые 20 INFO/WARN за секунду — всегда
    Thereafter: 50,   // потом каждое 50-е
    AlwaysPass: []lk.Level{lk.LevelError, lk.LevelFatal},
})
```

### 5.3 Дочерние логгеры с контекстом запроса

Можно создавать логгер с постоянными полями для каждого запроса:

```go
// В loggingMiddleware handler.go:
reqLogger := logger.KimLogger.With(
    "request_id", c.GetHeader("X-Request-ID"),
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
)
reqLogger.Info("Запрос принят")
// ... обработка ...
reqLogger.Info("Запрос завершён", "status", c.Writer.Status(), "duration", duration)
```

### 5.4 Отправка логов в OpenSearch (для production-мониторинга)

```go
osHandler, err := lk.NewOpenSearchHandlerWithConfig(lk.OpenSearchConfig{
    URL:           "http://opensearch:9200",
    Index:         "employee-logs",
    BufferSize:    100,
    FlushInterval: 5 * time.Second,
    Async:         true,
    Level:         lk.LevelWarn, // только WARN и выше → в OpenSearch
})

cfg := lk.Config{
    Outputs: []lk.Handler{stdoutHandler, fileHandler, osHandler},
    // ...
}
```

### 5.5 Динамическая смена уровня логирования

`logger_kim` поддерживает смену уровня без перезапуска приложения:

```go
// Переключить в DEBUG режим (например, при диагностике в production)
logger.KimLogger.SetLevel(lk.LevelDebug)

// Вернуть обратно
logger.KimLogger.SetLevel(lk.LevelInfo)
```

---

## 6. Схема работы после внедрения

```
Весь существующий код (handler.go, service, telemetry)
         │
         │  slog.Info("...")
         │  slog.Debug("...")
         │  slog.Error("...")
         ▼
    log/slog (stdlib)
         │
         │  SetDefault() через SlogBridge
         ▼
   logger_kim.kimLogger
         │
    ┌────┴────────────────────────┐
    ▼                             ▼
StdoutHandler                FileHandler
(FormatPretty)               (FormatJSON)
Цветной вывод             Файл с ротацией
в терминал                logs/app.log
```

---

## 7. Сводная таблица изменений

| Файл | Тип изменения | Описание |
|------|--------------|----------|
| `go.mod` | Обязательно | Добавить `require` + `replace` для `logger_kim` |
| `internal/logger/logger.go` | Обязательно | Полная замена реализации на `logger_kim.New()` |
| `cmd/server/main.go` | Обязательно | Изменить сигнатуру вызова `Setup()` |
| `internal/handler/handler.go` | Не нужно | `slog.*` вызовы работают через SlogBridge |
| `internal/service/employee_service.go` | Не нужно | `slog.DebugContext` работает через SlogBridge |
| `internal/telemetry/metrics.go` | Не нужно | `slog.Error` работает через SlogBridge |

---

## 8. Проверка после внедрения

```bash
# Запуск проекта
go run ./cmd/server/

# Ожидаемый вывод в терминале (Pretty с цветами):
# 2026-04-19 12:00:00 [INFO]  Логгер инициализирован  log_file=logs/app.log
# 2026-04-19 12:00:00 [INFO]  Запуск сервера  port=:8080

# В файле logs/app.log — чистый JSON:
# {"@timestamp":"2026-04-19T12:00:00Z","level":"INFO","message":"Логгер инициализирован","log_file":"logs/app.log"}
```

---

## 9. Возможные проблемы и решения

### Проблема: `go mod tidy` не находит `logger_kim`

**Причина:** Отсутствует `replace`-директива в `go.mod`.

**Решение:** Убедиться, что в `go.mod` основного проекта добавлено:
```
replace github.com/metal-telo/Logger-na-GO-KIP/logger_kim => ./logger_kim
```

### Проблема: Несовместимость версий Go

`logger_kim/go.mod` объявляет `go 1.25.1`, основной проект — `go 1.21`. При `go mod tidy` может появиться предупреждение.

**Решение:** Обновить `go.mod` основного проекта до `go 1.21` или выше; при необходимости скорректировать версию в `logger_kim/go.mod`.

### Проблема: Потеря логов при аварийном завершении

При использовании `AsyncHandler` буфер может не успеть сброситься.

**Решение:** Обработка сигналов (`SIGTERM`, `SIGINT`) уже реализована в `cmd/server/main.go` через `signal.Notify`. `defer closeLogger()` будет вызван корректно через `run()`.

### Проблема: `slog.DebugContext` не выводит логи

`logger_kim` по умолчанию устанавливает `LevelInfo`. Debug-сообщения будут отброшены.

**Решение:** В конфиге установить `Level: lk.LevelDebug` — либо в `Config`, либо вызвав `KimLogger.SetLevel(lk.LevelDebug)` динамически.

---

## 10. Итоговое заключение

`logger_kim` спроектирована именно для такого сценария внедрения: проект уже использует `log/slog`, и заменять все вызовы вручную — лишняя работа. Механизм `SlogBridge` позволяет внедрить библиотеку **изменив только 3 файла** (`go.mod`, `internal/logger/logger.go`, `cmd/server/main.go`), получив при этом:

- Цветной Pretty-вывод при разработке
- JSON-файл с автоматической ротацией
- Возможность асинхронной записи
- Динамическую смену уровня логирования
- Семплирование при высокой нагрузке
- Интеграцию с OpenSearch для централизованного мониторинга

---

## 11. Полная интеграция с OpenSearch: пошаговая инструкция

### Что такое OpenSearch в контексте этого проекта

OpenSearch — это поисковый движок (форк Elasticsearch), куда `logger_kim` отправляет логи через **Bulk API** пакетами. После этого можно искать, фильтровать и визуализировать логи через **OpenSearch Dashboards** (аналог Kibana).

Данные, которые будет получать OpenSearch из этого проекта:
- HTTP-запросы: метод, путь, статус, длительность, IP клиента
- Debug-события из бизнес-логики: создание/обновление сотрудников
- Ошибки API и ошибки метрик
- Системные события: запуск/остановка сервера

---

### Шаг 11.1: Установка OpenSearch через Docker

Создать файл `docker-compose.yml` в корне проекта `Logger-na-GO-KIP/`:

```yaml
version: "3.8"

services:
  opensearch:
    image: opensearchproject/opensearch:2.13.0
    container_name: opensearch
    environment:
      - discovery.type=single-node
      - DISABLE_SECURITY_PLUGIN=true       # отключает TLS/auth для локальной разработки
      - OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m
    ports:
      - "9200:9200"    # REST API — сюда шлёт логи logger_kim
      - "9600:9600"    # Performance Analyzer
    volumes:
      - opensearch-data:/usr/share/opensearch/data

  opensearch-dashboards:
    image: opensearchproject/opensearch-dashboards:2.13.0
    container_name: opensearch-dashboards
    ports:
      - "5601:5601"    # Веб-интерфейс для просмотра логов
    environment:
      - OPENSEARCH_HOSTS=["http://opensearch:9200"]
      - DISABLE_SECURITY_DASHBOARDS_PLUGIN=true
    depends_on:
      - opensearch

volumes:
  opensearch-data:
```

Запустить:

```bash
# Из директории Logger-na-GO-KIP/
docker compose up -d

# Проверить, что OpenSearch поднялся (ждать ~30 секунд)
curl http://localhost:9200/_cluster/health
# Ожидаемый ответ: {"status":"green"} или {"status":"yellow"}

# Открыть дашборды в браузере
# http://localhost:5601
```

> **Системные требования:** Docker Desktop для Windows. Минимум 4 GB RAM (OpenSearch требовательна к памяти).

---

### Шаг 11.2: Как работает `OpenSearchHandler` внутри `logger_kim`

При создании `OpenSearchHandler` происходит следующее:

1. **`healthCheck()`** — GET `/_cluster/health` → убеждается что OpenSearch доступен
2. **`createIndex()`** — HEAD `/employee-logs` → если индекс не существует, создаёт его с маппингом:
   ```json
   {
     "mappings": {
       "properties": {
         "@timestamp": { "type": "date" },
         "level":      { "type": "keyword" },
         "message":    { "type": "text" }
       }
     }
   }
   ```
3. **`flushLoop()`** — горутина, которая каждые `FlushInterval` секунд отправляет буфер через **Bulk API** (`POST /_bulk`)

Формат одного документа в OpenSearch (как его видит Dashboards):

```json
{
  "@timestamp": "2026-04-19T12:34:56.789Z",
  "level": "INFO",
  "message": "HTTP request",
  "method": "POST",
  "path": "/api/employees",
  "status": 201,
  "duration": "12.5ms",
  "client_ip": "127.0.0.1",
  "service": "employee-management"
}
```

---

### Шаг 11.3: Подключение `OpenSearchHandler` в проекте

**Обновить `internal/logger/logger.go`** — добавить OpenSearch как третий output:

```go
package logger

import (
    "fmt"
    "os"
    "time"

    lk "github.com/metal-telo/Logger-na-GO-KIP/logger_kim"
)

var KimLogger lk.Logger

func Setup(logDir, logFile string) (func(), error) {
    logFilePath := logDir + "/" + logFile

    // Handler 1: цветной Pretty-вывод в терминал
    stdoutHandler := lk.NewStdoutHandlerWithFormat(lk.FormatPretty)

    // Handler 2: JSON-файл с ротацией
    fileHandler, err := lk.NewFileHandlerWithConfig(lk.FileHandlerConfig{
        FilePath:   logFilePath,
        MaxSize:    100,
        MaxBackups: 5,
        Format:     lk.FormatJSON,
        Level:      lk.LevelInfo,
    })
    if err != nil {
        return nil, fmt.Errorf("не удалось создать FileHandler: %w", err)
    }

    // Handler 3: OpenSearch (только WARN и выше, чтобы не перегружать)
    opensearchURL := os.Getenv("OPENSEARCH_URL")
    if opensearchURL == "" {
        opensearchURL = "http://localhost:9200"
    }
    osHandler, err := lk.NewOpenSearchHandlerWithConfig(lk.OpenSearchConfig{
        URL:           opensearchURL,
        Index:         "employee-logs",
        BufferSize:    50,                  // накопить 50 записей перед отправкой
        FlushInterval: 5 * time.Second,     // или сбрасывать каждые 5 секунд
        Async:         true,                // не блокировать основной поток
        Timeout:       10 * time.Second,
        Level:         lk.LevelWarn,        // в OpenSearch: только WARN, ERROR, FATAL
    })
    if err != nil {
        // OpenSearch недоступен — продолжаем без него, не падаем
        fmt.Printf("[WARN] OpenSearch недоступен, логи идут только в stdout+file: %v\n", err)
        osHandler = nil
    }

    outputs := []lk.Handler{stdoutHandler, fileHandler}
    if osHandler != nil {
        outputs = append(outputs, osHandler)
    }

    cfg := lk.Config{
        Level:       lk.LevelDebug,
        ServiceName: "employee-management",
        Outputs:     outputs,
        Features: lk.Features{
            SlogBridge:       true,
            PrettyColor:      true,
            StructuredFields: true,
            MultipleOutputs:  true,
            FileOutput:       true,
            OpenSearchOutput: true,
        },
    }

    KimLogger = lk.New(cfg)
    return func() { _ = KimLogger.Close() }, nil
}
```

> **Важно:** если OpenSearch недоступен при старте — `NewOpenSearchHandlerWithConfig` вернёт ошибку из `healthCheck()`. В коде выше это обрабатывается мягко: сервер запускается без OpenSearch.

---

### Шаг 11.4: Переменные окружения

Создать файл `.env` в корне `Logger-na-GO-KIP/`:

```env
# OpenSearch
OPENSEARCH_URL=http://localhost:9200

# Приложение
APP_PORT=:8080
LOG_LEVEL=debug
```

Добавить чтение `.env` в `cmd/server/main.go` (через `os.Getenv` — уже используется в `Setup()`).

---

### Шаг 11.5: Запуск всего стека

```bash
# 1. Поднять OpenSearch
docker compose up -d

# 2. Дождаться готовности (~30 сек)
curl http://localhost:9200/_cluster/health?pretty

# 3. Запустить Go-приложение
go run ./cmd/server/

# 4. Создать тестовые данные — сделать несколько запросов
curl -X POST http://localhost:8080/api/employees \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Иван Иванов","position_id":1,"department_id":1}'

# 5. Проверить, что логи попали в OpenSearch
curl "http://localhost:9200/employee-logs/_search?pretty&size=5"
```

Ответ OpenSearch должен содержать документы вида:

```json
{
  "_source": {
    "@timestamp": "2026-04-19T12:34:56Z",
    "level": "WARN",
    "message": "...",
    "service": "employee-management"
  }
}
```

---

### Шаг 11.6: Просмотр логов в OpenSearch Dashboards

Открыть в браузере: **http://localhost:5601**

#### Создать Data View (Index Pattern)

1. Перейти: `Management → Stack Management → Index Patterns`
2. Нажать `Create index pattern`
3. В поле `Index pattern name` ввести: `employee-logs*`
4. В поле `Time field` выбрать: `@timestamp`
5. Нажать `Create index pattern`

#### Открыть Discover

1. Перейти: `Discover` (в левом меню)
2. Выбрать Data View: `employee-logs*`
3. Логи из проекта появятся в хронологическом порядке

#### Полезные фильтры в Discover

| Что найти | KQL-запрос в строке поиска |
|-----------|---------------------------|
| Только ошибки | `level: "ERROR"` |
| Только предупреждения и выше | `level: "WARN" or level: "ERROR"` |
| Запросы к конкретному пути | `path: "/api/employees"` |
| Медленные запросы (если duration в числах) | `duration > 100` |
| Все логи за последний час | Использовать фильтр времени вверху справа |

---

### Шаг 11.7: Создание дашборда в OpenSearch Dashboards

1. Перейти: `Dashboard → Create new dashboard`
2. Нажать `Add visualization`
3. Примеры полезных визуализаций:

**График: количество запросов по времени**
- Тип: `Vertical bar`
- X-ось: `@timestamp` (Date histogram, интервал Auto)
- Y-ось: `Count`

**Пирог: распределение по уровням логов**
- Тип: `Pie`
- Разбивка: поле `level` (Terms aggregation)

**Таблица: последние ошибки**
- Тип: `Data Table`
- Фильтр: `level: "ERROR"`
- Колонки: `@timestamp`, `message`, `path`

---

### Шаг 11.8: Что происходит при отправке — технические детали

```
Go-приложение (slog.Warn / slog.Error)
         │
         ▼
  logger_kim.kimLogger
         │  (через SlogBridge)
         ▼
  OpenSearchHandler.Handle()
         │
         │  entry → буфер ([]Entry)
         ▼
    Буфер заполнен (50 записей)  ──или──  flushTicker (5 сек)
         │
         ▼
  bulkSend() → POST http://localhost:9200/_bulk
         │
         │  NDJSON-тело:
         │  {"index":{"_index":"employee-logs"}}
         │  {"@timestamp":"...","level":"WARN","message":"..."}
         │  {"index":{"_index":"employee-logs"}}
         │  {"@timestamp":"...","level":"ERROR","message":"..."}
         │
         ▼
     OpenSearch сохраняет документы
         │
         ▼
  Доступно в Dashboards (http://localhost:5601)
```

---

### Шаг 11.9: Остановка и очистка

```bash
# Остановить контейнеры (данные сохраняются в volume)
docker compose down

# Остановить и удалить все данные
docker compose down -v

# Посмотреть список индексов в OpenSearch
curl http://localhost:9200/_cat/indices?v

# Удалить индекс с логами (осторожно — необратимо)
curl -X DELETE http://localhost:9200/employee-logs
```

---

## 12. Для новичков: суть работы с `logger_kim` и OpenSearch простыми словами

### Что такое лог и зачем он нужен

Представь, что твоё Go-приложение — это кухня в ресторане. Повара работают, готовят блюда, иногда что-то идёт не так. **Лог** — это журнал, куда записывается всё, что происходит: «принят заказ», «закончилась соль», «сгорела котлета».

Без логов, когда что-то сломается, ты не будешь знать — когда, где и почему. С логами — открываешь журнал и смотришь.

```
Без логов:  «Сервер упал... почему? когда? что случилось?» 😱
С логами:   «В 14:32 пришёл запрос POST /api/employees, вернул ошибку 500, 
             причина: переполнение буфера» 😌
```

---

### Что такое `log/slog` и зачем `logger_kim`

**`log/slog`** — это стандартная библиотека Go для логирования. Она уже есть в проекте. Простая, но ограниченная:

```go
slog.Info("Пользователь создан", "id", 42)
// Выводит: {"time":"...","level":"INFO","msg":"Пользователь создан","id":42}
```

**`logger_kim`** — это расширение над `slog`. Она не заменяет `slog`, а даёт ему суперспособности:

| Что умеет | `slog` (стандартный) | `logger_kim` |
|-----------|----------------------|--------------|
| Писать в файл | Только вручную | Встроено + ротация |
| Цветной вывод в терминал | Нет | Да (Pretty) |
| Отправлять в OpenSearch | Нет | Да (Bulk API) |
| Асинхронная запись | Нет | Да (AsyncHandler) |
| Уровень без перезапуска | Нет | Да (SetLevel) |
| Работать с существующим slog-кодом | — | Да (SlogBridge) |

**Самое важное:** после внедрения `logger_kim` твой старый код `slog.Info(...)`, `slog.Error(...)` **не нужно переписывать**. Всё продолжает работать, но теперь логи идут ещё и в файл, и в OpenSearch.

---

### Что такое OpenSearch и зачем он нужен

**OpenSearch** — это программа, которая:
1. **Принимает** данные (в нашем случае — логи) через HTTP
2. **Хранит** их в виде документов (как строки в базе данных, но очень быстро ищет по тексту)
3. **Позволяет искать** по любому полю: «покажи все ERROR за вчера» или «покажи запросы длиннее 500мс»
4. **Рисует графики** через веб-интерфейс Dashboards

Аналогия: `slog` пишет лог в тетрадь. `logger_kim` — умная ручка, которая одновременно пишет в тетрадь, в файл и отправляет копию в электронную базу. OpenSearch — это та самая электронная база с поиском и графиками.

```
Твой код            logger_kim              Куда попадает
─────────           ──────────              ─────────────
slog.Info(...)  →   StdoutHandler     →     Терминал (с цветами)
                    FileHandler       →     Файл logs/app.log
                    OpenSearchHandler →     OpenSearch (поиск + графики)
```

---

### Почему логи отправляются «пачками», а не по одному

Представь: каждый раз, когда ты записываешь одно слово в Word, он сохраняет файл на диск. Это очень медленно. Поэтому Word накапливает изменения и сохраняет раз в несколько секунд.

`OpenSearchHandler` работает так же — это называется **буферизация**:

```
Лог 1 → буфер [1]
Лог 2 → буфер [1, 2]
Лог 3 → буфер [1, 2, 3]
...
Лог 50 → буфер заполнен (50 штук)!
         ИЛИ прошло 5 секунд
         → отправить все 50 за один HTTP-запрос → OpenSearch
```

Это называется **Bulk API** — «пакетная» отправка. Вместо 50 HTTP-запросов — один. Быстрее в 10-50 раз.

```go
// В OpenSearchConfig это настраивается:
BufferSize:    50,               // накопить 50 записей
FlushInterval: 5 * time.Second, // или сбросить через 5 секунд
```

---

### Что значат уровни логов

Уровень лога — это «срочность» сообщения. В `logger_kim` их 5:

| Уровень | Когда использовать | Пример |
|---------|-------------------|--------|
| `DEBUG` | Детали для разработчика, в production обычно выключен | `"Запрос к БД: SELECT * FROM..."` |
| `INFO` | Обычные события, всё работает нормально | `"Сервер запущен на :8080"` |
| `WARN` | Что-то необычное, но не сломано | `"Высокая нагрузка CPU: 85%"` |
| `ERROR` | Что-то сломалось, но сервер продолжает работать | `"Ошибка создания сотрудника"` |
| `FATAL` | Критическая ошибка, приложение не может продолжать | `"Не удалось подключиться к БД"` |

В этом проекте `OpenSearchHandler` настроен на `LevelWarn` — это значит, что в OpenSearch попадают **только WARN, ERROR и FATAL**. DEBUG и INFO идут только в терминал и файл. Это разумно: в OpenSearch хранить миллион «всё хорошо» — пустая трата места.

---

### Что такое `SlogBridge` и почему это важно для новичка

Это **самая магическая часть** `logger_kim`. Объясняю на пальцах:

Допустим, ты купил умную розетку, которая считает потребление электричества. Чтобы все приборы в квартире автоматически считались — ты вставляешь её в щиток, и теперь всё электричество идёт через неё. Ты не перематываешь провода у каждого прибора.

`SlogBridge` делает то же самое:

```go
// Было: стандартный slog пишет в JSON в файл
slog.SetDefault(slog.New(slog.NewJSONHandler(file, nil)))

// Стало: logger_kim перехватывает все вызовы slog
Features: lk.Features{
    SlogBridge: true,  // ← все slog.* теперь идут через logger_kim
}
```

После этого любой код в проекте, который пишет `slog.Info(...)` — автоматически получает Pretty-вывод, запись в файл и отправку в OpenSearch. **Без изменения этого кода.**

---

### Жизненный цикл одного лога: от кода до Dashboards

Разберём на примере: что происходит, когда кто-то создаёт сотрудника через API.

```
1. Клиент делает запрос:
   POST /api/employees {"full_name": "Иван", ...}

2. handler.go обрабатывает запрос и пишет лог:
   slog.Info("HTTP request", "method", "POST", "path", "/api/employees", "status", 201)

3. SlogBridge перехватывает этот вызов и передаёт в logger_kim

4. logger_kim создаёт Entry:
   Entry{
     Time:    2026-04-19T12:34:56Z,
     Level:   INFO,
     Message: "HTTP request",
     Fields:  {"method":"POST", "path":"/api/employees", "status":201}
   }

5. Entry рассылается в три Handler-а параллельно:
   ├── StdoutHandler → цветная строка в терминал
   ├── FileHandler   → JSON-строка в logs/app.log
   └── OpenSearchHandler → Entry добавляется в буфер (INFO < WARN, поэтому пропускается)

6. service.go замечает что-то подозрительное:
   slog.Warn("Сотрудник уже существует", "name", "Иван")
   → Entry{Level: WARN} → OpenSearchHandler принимает (WARN >= WARN)
   → буфер: [entry1, entry2, ...]
   → когда буфер заполнен или прошло 5 сек → bulkSend()

7. bulkSend() отправляет HTTP POST /_bulk в OpenSearch:
   {"index":{"_index":"employee-logs"}}
   {"@timestamp":"2026-04-19T12:34:56Z","level":"WARN","message":"Сотрудник уже существует","name":"Иван"}

8. OpenSearch сохраняет документ в индекс "employee-logs"

9. Разработчик открывает http://localhost:5601
   → Discover → фильтр level: "WARN"
   → видит это событие с полным контекстом
```

---

### Частые вопросы новичков

**В: Если OpenSearch не запущен, сервер упадёт?**

О: Нет, если написать обработку ошибки как показано в Шаге 11.3 — сервер запустится без OpenSearch. Логи будут идти только в терминал и файл.

**В: Зачем файл, если есть OpenSearch?**

О: Файл — это страховка. OpenSearch может быть недоступен, может переполниться, может сломаться. Файл всегда на месте. Кроме того, файл удобно читать через `tail -f logs/app.log` прямо в терминале.

**В: Почему DEBUG не отправляется в OpenSearch?**

О: DEBUG-сообщений очень много (каждый запрос к БД, каждая итерация цикла). Если отправлять всё — OpenSearch быстро заполнится. В OpenSearch имеет смысл хранить только то, на что нужно обращать внимание: предупреждения и ошибки.

**В: Нужно ли платить за OpenSearch?**

О: Нет. OpenSearch — open source, полностью бесплатный. Docker-образ скачивается бесплатно.

**В: Что будет, если приложение аварийно завершится, а в буфере есть непосланные логи?**

О: Они потеряются. Это компромисс производительности. Для критически важных логов используй `LevelError` с маленьким `BufferSize: 1` или `FlushInterval: 1*time.Second`, чтобы они уходили почти мгновенно.
