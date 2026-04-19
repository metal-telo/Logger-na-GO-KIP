markdown
# Logger KIM (Keep It Meaningful)

[![Go Reference](https://pkg.go.dev/badge/github.com/metal-telo/Logger-na-GO-KIP/logger_kim.svg)](https://pkg.go.dev/github.com/metal-telo/Logger-na-GO-KIP/logger_kim)
[![Go Report Card](https://goreportcard.com/badge/github.com/metal-telo/Logger-na-GO-KIP/logger_kim)](https://goreportcard.com/report/github.com/metal-telo/Logger-na-GO-KIP/logger_kim)

**Logger KIM** — гибкая, интерфейсная библиотека структурированного логирования для Go, построенная поверх `log/slog` (stdlib) и расширяющая его возможности.

## Возможности

- **Plug-in архитектура Handler-ов** — stdout, stderr, file, OpenSearch, null, multi
- **Цветной Pretty-вывод** для разработки (на базе `github.com/fatih/color`)
- **Ротация файлов** по размеру и по дате — встроена, без внешних зависимостей
- **Bulk-буферизация для OpenSearch** — использует Bulk API, а не per-record HTTP
- **Динамическая смена уровня** — `SetLevel()` / `GetLevel()` без перезапуска
- **Контекстные дочерние логгеры** — `With()` / `WithContext()`
- **Глобальный логгер по умолчанию** — `SetDefault()` / `Default()`
- **Асинхронная запись** — `AsyncHandler` устраняет блокировки на I/O
- **Семплирование** — `SamplingHandler` и `ProbabilisticHandler` для высоконагруженных систем
- **Slog Bridge** — полная интеграция с `log/slog` и стандартным пакетом `log`
- **Декларативные Features** — включение/отключение возможностей через конфиг

## Установка

```bash
go get github.com/metal-telo/Logger-na-GO-KIP/logger_kim
Быстрый старт
1. Простой пример
go
package main

import (
    "github.com/metal-telo/Logger-na-GO-KIP/logger_kim"
)

func main() {
    // Создаём логгер с конфигурацией по умолчанию (JSON в stdout)
    log := logger_kim.New(logger_kim.DefaultConfig())
    defer log.Close()

    log.Info("Сервер запущен", "port", 8080, "env", "development")
    log.Debug("Детали подключения", "host", "localhost")
    log.Warn("Высокая загрузка", "cpu", 85.5)
    log.Error("Ошибка БД", "error", "connection refused")
}
2. Pretty-вывод для разработки
go
cfg := logger_kim.Config{
    Level:  logger_kim.LevelDebug,
    Format: logger_kim.FormatPretty,
    Outputs: []logger_kim.Handler{
        logger_kim.NewStdoutHandlerWithFormat(logger_kim.FormatPretty),
    },
    Features: logger_kim.DefaultFeatures(),
}
log := logger_kim.New(cfg)
3. Запись в файл с ротацией
go
fileHandler, _ := logger_kim.NewFileHandlerWithConfig(logger_kim.FileHandlerConfig{
    FilePath:   "./logs/app.log",
    MaxSize:    100,  // 100 MB
    MaxBackups: 5,
    Format:     logger_kim.FormatJSON,
    Level:      logger_kim.LevelInfo,
})

cfg := logger_kim.Config{
    Level:   logger_kim.LevelInfo,
    Outputs: []logger_kim.Handler{fileHandler},
}
log := logger_kim.New(cfg)
4. Отправка логов в OpenSearch
go
osHandler, err := logger_kim.NewOpenSearchHandlerWithConfig(logger_kim.OpenSearchConfig{
    URL:           "http://localhost:9200",
    Index:         "app-logs",
    BufferSize:    100,
    FlushInterval: 5 * time.Second,
    Async:         true,
})
if err != nil {
    panic(err)
}

cfg := logger_kim.Config{
    Outputs: []logger_kim.Handler{osHandler},
}
log := logger_kim.New(cfg)
5. Production-конфигурация с Async и Sampling
go
cfg := logger_kim.Config{
    Level:    logger_kim.LevelInfo,
    Format:   logger_kim.FormatJSON,
    Features: logger_kim.ProductionFeatures(),
    Outputs: []logger_kim.Handler{
        logger_kim.NewStdoutHandler(),
        fileHandler,
    },
}
log := logger_kim.New(cfg)
6. Интеграция с log/slog
go
kimLog := logger_kim.New(cfg)
logger_kim.SetAsDefault(kimLog) // slog.Info / log.Println идут через Logger KIM

slog.Info("Привет из slog", "user", "admin")
log.Println("Старый добрый log.Printf тоже работает")
7. Глобальные функции
go
logger_kim.SetDefault(log)
logger_kim.Info("Глобальное сообщение")
logger_kim.With("request_id", "abc-123").Info("С контекстом")
8. Создание дочерних логгеров
go
requestLogger := log.With("request_id", uuid.New().String())
requestLogger.Info("Начало обработки запроса")

userLogger := log.WithContext(ctx).With("user_id", 42)
userLogger.Info("Пользователь авторизован")
Архитектура
Logger KIM построен вокруг трёх ключевых интерфейсов:

text
┌─────────────────────────────────────────────────────────────┐
│                         Logger                              │
│  (основной интерфейс: Debug, Info, Warn, Error, Fatal)     │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                        Handlers                             │
│  (stdout, file, OpenSearch, async, sampling, multi...)     │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Formatters                             │
│  (JSON, Text, Pretty, Logfmt, Custom)                      │
└─────────────────────────────────────────────────────────────┘
Порядок оборачивания Handlers
Когда включены Features.Async и Features.Sampling, обёртки применяются в порядке:

text
SamplingHandler → AsyncHandler → Реальный Handler (File/Stdout/OpenSearch)
Доступные Handlers
Handler	Назначение
StdoutHandler	Вывод в stdout
StderrHandler	Вывод в stderr
FileHandler	Запись в файл с ротацией по размеру
DailyRotatingFileHandler	Ротация по дате
OpenSearchHandler	Отправка в OpenSearch/Elasticsearch через Bulk API
AsyncHandler	Асинхронная обёртка над любым handler
SamplingHandler	Семплирование по счётчику за период
ProbabilisticHandler	Вероятностное семплирование
MultiHandler	Запись в несколько handlers одновременно
NullHandler	Отключение вывода
Доступные Formatters
Formatter	Назначение
JSONFormatter	Структурированный JSON
PrettyJSONFormatter	Pretty-printed JSON
TextFormatter	Простой текстовый формат
PrettyFormatter	Цветной вывод для разработки
LogfmtFormatter	Формат logfmt (key=value)
CustomFormatter	Пользовательская функция форматирования
Features — декларативное управление возможностями
go
type Features struct {
    Async            bool          // Асинхронная запись
    AsyncOpts        AsyncOptions  // Параметры AsyncHandler
    Sampling         bool          // Семплирование
    SamplingOpts     SamplingConfig // Параметры семплирования
    SlogBridge       bool          // Автоматический SetAsDefault()
    PrettyColor      bool          // Цвета в Pretty-формате
    ColorScheme      ColorScheme   // Кастомная цветовая схема
    StructuredFields bool          // Вывод структурированных полей
    MultipleOutputs  bool          // Разрешить несколько handlers
    FileOutput       bool          // Включить FileHandler
    OpenSearchOutput bool          // Включить OpenSearchHandler
}
Готовые пресеты:

DefaultFeatures() — базовые возможности, без async/sampling

MinimalFeatures() — минимальный логгер для CLI/тестов

ProductionFeatures() — async + sampling + slog bridge