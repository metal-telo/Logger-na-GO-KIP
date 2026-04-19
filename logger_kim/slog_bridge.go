package logger_kim

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"strings"
	"time"
)

// ─── SlogHandler ─────────────────────────────────────────────────────────────
//
// SlogHandler реализует интерфейс slog.Handler поверх logger_kim.Logger.
// Это позволяет использовать logger_kim как бэкенд для slog:
//
//	kimLog := logger_kim.New(cfg)
//	slog.SetDefault(slog.New(logger_kim.NewSlogHandler(kimLog)))
//
// После этого все вызовы slog.Info / slog.Debug / ... проходят через
// logger_kim — получают pretty-вывод, async, sampling, OpenSearch и т.д.
type SlogHandler struct {
	logger Logger
	attrs  []slog.Attr  // накопленные WithAttrs
	group  string       // текущий префикс группы
}

// NewSlogHandler создаёт slog.Handler, делегирующий в logger_kim.Logger.
func NewSlogHandler(l Logger) *SlogHandler {
	return &SlogHandler{logger: l}
}

// Enabled проверяет, активен ли уровень в logger_kim.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch {
	case level >= slog.LevelError:
		return true // Error/Fatal всегда
	case level >= slog.LevelWarn:
		return true
	default:
		// Делегировать точную проверку logger_kim невозможно без доступа к
		// его уровню через интерфейс — поэтому разрешаем и даём ему решать.
		return true
	}
}

// Handle конвертирует slog.Record в logger_kim.Entry и передаёт логгеру.
func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	args := make([]any, 0, r.NumAttrs()*2+len(h.attrs)*2)

	// 1. Предварительно накопленные поля (WithAttrs)
	for _, a := range h.attrs {
		key := h.prefixKey(a.Key)
		args = append(args, key, a.Value.Any())
	}

	// 2. Поля текущей записи
	r.Attrs(func(a slog.Attr) bool {
		key := h.prefixKey(a.Key)
		args = append(args, key, a.Value.Any())
		return true
	})

	msg := r.Message
	switch {
	case r.Level >= slog.LevelError:
		h.logger.WithContext(ctx).Error(msg, args...)
	case r.Level >= slog.LevelWarn:
		h.logger.WithContext(ctx).Warn(msg, args...)
	case r.Level >= slog.LevelInfo:
		h.logger.WithContext(ctx).Info(msg, args...)
	default:
		h.logger.WithContext(ctx).Debug(msg, args...)
	}
	return nil
}

// WithAttrs возвращает новый handler с дополнительными постоянными полями.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	copy(merged[len(h.attrs):], attrs)
	return &SlogHandler{logger: h.logger, attrs: merged, group: h.group}
}

// WithGroup возвращает новый handler, добавляющий префикс "group." к ключам.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	prefix := name
	if h.group != "" {
		prefix = h.group + "." + name
	}
	return &SlogHandler{logger: h.logger, attrs: h.attrs, group: prefix}
}

func (h *SlogHandler) prefixKey(key string) string {
	if h.group == "" {
		return key
	}
	return h.group + "." + key
}

// ─── SetAsDefault ─────────────────────────────────────────────────────────────
//
// SetAsDefault регистрирует logger_kim как глобальный бэкенд для двух вещей:
//  1. slog.SetDefault — все slog.Info / slog.Warn / ... идут через logger_kim
//  2. log.SetOutput — старый log.Println / log.Printf тоже идут через logger_kim
//
// Вызвать один раз в main() после инициализации логгера.
func SetAsDefault(l Logger) {
	// 1. slog default
	slog.SetDefault(slog.New(NewSlogHandler(l)))

	// 2. Перехватить старый log пакет через кастомный Writer
	log.SetOutput(&stdBridgeWriter{logger: l})
	log.SetFlags(0) // убираем дублирующийся timestamp от stdlib log
}

// stdBridgeWriter направляет вывод stdlib log → logger_kim.Info
type stdBridgeWriter struct {
	logger Logger
}

func (w *stdBridgeWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(bytes.TrimSpace(p)), "\n")
	if msg != "" {
		w.logger.Info(msg, "source", "stdlib/log")
	}
	return len(p), nil
}

// ─── FromSlog ─────────────────────────────────────────────────────────────────
//
// FromSlog адаптирует *slog.Logger обратно в logger_kim.Logger — полезно при
// постепенной migration: функции, принимающие logger_kim.Logger, могут
// получить обёртку поверх существующего *slog.Logger.
type slogToKim struct {
	sl *slog.Logger
}

// FromSlog создаёт logger_kim.Logger, делегирующий вызовы в *slog.Logger.
func FromSlog(sl *slog.Logger) Logger {
	return &slogToKim{sl: sl}
}

func (s *slogToKim) Debug(msg string, args ...any) { s.sl.Debug(msg, args...) }
func (s *slogToKim) Info(msg string, args ...any)  { s.sl.Info(msg, args...) }
func (s *slogToKim) Warn(msg string, args ...any)  { s.sl.Warn(msg, args...) }
func (s *slogToKim) Error(msg string, args ...any) { s.sl.Error(msg, args...) }
func (s *slogToKim) Fatal(msg string, args ...any) {
	s.sl.Error("[FATAL] "+msg, args...)
	panic(msg)
}
func (s *slogToKim) With(args ...any) Logger {
	return &slogToKim{sl: s.sl.With(args...)}
}
func (s *slogToKim) WithContext(ctx context.Context) Logger {
	return &slogToKim{sl: s.sl.With("ctx", ctx)}
}
func (s *slogToKim) SetLevel(_ Level)   {}
func (s *slogToKim) GetLevel() Level    { return LevelDebug }
func (s *slogToKim) Close() error       { return nil }

// Проверка времени компиляции: slogToKim удовлетворяет Logger
var _ Logger = (*slogToKim)(nil)

// Проверка времени компиляции: SlogHandler удовлетворяет slog.Handler
var _ slog.Handler = (*SlogHandler)(nil)

// ─── Helpers ─────────────────────────────────────────────────────────────────

// NewSlogLogger создаёт *slog.Logger, использующий logger_kim как бэкенд.
// Краткая форма: slog.New(logger_kim.NewSlogHandler(l))
func NewSlogLogger(l Logger) *slog.Logger {
	return slog.New(NewSlogHandler(l))
}

// TimeFromRecord возвращает time.Time из slog.Record, с fallback на now.
func TimeFromRecord(r slog.Record) time.Time {
	if !r.Time.IsZero() {
		return r.Time
	}
	return time.Now()
}
