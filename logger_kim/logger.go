// Package logger_kim — гибкая, интерфейсная библиотека структурированного логирования.
//
// Построена поверх log/slog (stdlib), расширяет его:
//   - Plug-in архитектура Handler-ов (stdout / file / OpenSearch / null)
//   - Цветной Pretty-вывод для разработки (github.com/fatih/color)
//   - Ротация файлов по размеру и по дате — встроена
//   - Bulk-буферизация для OpenSearch (Bulk API, не per-record HTTP)
//   - Динамическая смена уровня: SetLevel() / GetLevel()
//   - Контекстные дочерние логгеры: With() / WithContext()
//   - Глобальный логгер по умолчанию: SetDefault() / Default()
package logger_kim

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Logger — основной интерфейс
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)

	// With возвращает новый логгер с постоянными полями-контекстом
	With(args ...any) Logger

	// WithContext возвращает новый логгер, привязанный к context.Context
	WithContext(ctx context.Context) Logger

	// SetLevel динамически меняет уровень без перезапуска
	SetLevel(level Level)
	GetLevel() Level

	// Close завершает работу и флашит все handlers
	Close() error
}

// Entry — одна запись в лог
type Entry struct {
	Time    time.Time
	Level   Level
	Message string
	Fields  Fields
	Context context.Context
}

// Fields — структурированные поля
type Fields map[string]any

// ─── Реализация ──────────────────────────────────────────────────────────────

type kimLogger struct {
	config   Config
	handlers []Handler
	mu       sync.RWMutex
	closed   bool
	ctx      context.Context
	attrs    []any
}

// New создаёт новый Logger с заданной конфигурацией.
// Если config.Features содержит флаги, они применяются автоматически:
// – Async/Sampling оборачивают все handlers
// – SlogBridge вызывает SetAsDefault(logger)
func New(config Config) Logger {
	if config.Level == 0 && config.Format == "" {
		config.Level = LevelInfo
	}
	if config.Format == "" {
		config.Format = FormatJSON
	}
	if len(config.Outputs) == 0 {
		config.Outputs = []Handler{NewStdoutHandler()}
	}

	// Применяем feature-флаги к handlers
	f := config.Features
	handlers := applyFeatures(config.Outputs, f)

	l := &kimLogger{
		config:   config,
		handlers: handlers,
		ctx:      context.Background(),
		attrs:    []any{},
	}

	// Автоматический slog-bridge если запрошен
	if f.SlogBridge {
		SetAsDefault(l)
	}

	return l
}

func (l *kimLogger) Debug(msg string, args ...any) { l.log(LevelDebug, msg, args...) }
func (l *kimLogger) Info(msg string, args ...any)  { l.log(LevelInfo, msg, args...) }
func (l *kimLogger) Warn(msg string, args ...any)  { l.log(LevelWarn, msg, args...) }
func (l *kimLogger) Error(msg string, args ...any) { l.log(LevelError, msg, args...) }
func (l *kimLogger) Fatal(msg string, args ...any) {
	l.log(LevelFatal, msg, args...)
	l.Close()
	panic(fmt.Sprintf("FATAL: %s", msg))
}

func (l *kimLogger) With(args ...any) Logger {
	newAttrs := make([]any, len(l.attrs)+len(args))
	copy(newAttrs, l.attrs)
	copy(newAttrs[len(l.attrs):], args)
	return &kimLogger{
		config:   l.config,
		handlers: l.handlers,
		ctx:      l.ctx,
		attrs:    newAttrs,
	}
}

func (l *kimLogger) WithContext(ctx context.Context) Logger {
	return &kimLogger{
		config:   l.config,
		handlers: l.handlers,
		ctx:      ctx,
		attrs:    l.attrs,
	}
}

func (l *kimLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

func (l *kimLogger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Level
}

func (l *kimLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	var lastErr error
	for _, h := range l.handlers {
		if err := h.Close(); err != nil {
			lastErr = err
		}
	}
	l.closed = true
	return lastErr
}

func (l *kimLogger) log(level Level, msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.closed || level < l.config.Level {
		return
	}

	allArgs := make([]any, 0, len(l.attrs)+len(args))
	allArgs = append(allArgs, l.attrs...)
	allArgs = append(allArgs, args...)

	entry := Entry{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: msg,
		Fields:  argsToFields(allArgs),
		Context: l.ctx,
	}

	for _, h := range l.handlers {
		if err := h.Handle(l.ctx, entry); err != nil {
			slog.Error("logger_kim: handler error",
				"handler", fmt.Sprintf("%T", h),
				"error", err,
			)
		}
	}
}

func argsToFields(args []any) Fields {
	fields := make(Fields, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key := fmt.Sprintf("%v", args[i])
		fields[key] = args[i+1]
	}
	return fields
}

// ─── Глобальный логгер ───────────────────────────────────────────────────────

var (
	defaultLogger   Logger
	defaultLoggerMu sync.RWMutex
)

func init() {
	defaultLogger = New(DefaultConfig())
}

func SetDefault(logger Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = logger
}

func Default() Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

func Debug(msg string, args ...any) { Default().Debug(msg, args...) }
func Info(msg string, args ...any)  { Default().Info(msg, args...) }
func Warn(msg string, args ...any)  { Default().Warn(msg, args...) }
func Error(msg string, args ...any) { Default().Error(msg, args...) }
func Fatal(msg string, args ...any) { Default().Fatal(msg, args...) }
func With(args ...any) Logger       { return Default().With(args...) }
