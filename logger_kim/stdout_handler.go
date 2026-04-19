package logger_kim

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// ─── StdoutHandler ───────────────────────────────────────────────────────────

type StdoutHandler struct {
	BaseHandler
	mu        sync.Mutex
	formatter Formatter
}

func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{
		BaseHandler: NewBaseHandler(LevelDebug, FormatJSON),
		formatter:   NewJSONFormatter(),
	}
}

func NewStdoutHandlerWithFormat(format Format) *StdoutHandler {
	var f Formatter
	switch format {
	case FormatText:
		f = NewTextFormatter()
	case FormatPretty:
		f = NewPrettyFormatter()
	case FormatLogfmt:
		f = NewLogfmtFormatter()
	default:
		f = NewJSONFormatter()
	}
	return &StdoutHandler{
		BaseHandler: NewBaseHandler(LevelDebug, format),
		formatter:   f,
	}
}

func (h *StdoutHandler) Handle(_ context.Context, entry Entry) error {
	if !h.Enabled(entry.Level) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	output, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, output)
	return nil
}

func (h *StdoutHandler) Close() error { return nil }

// SetFormatter заменяет форматтер StdoutHandler.
// Используется applyFeatures для подстановки ColorScheme из Features.
func (h *StdoutHandler) SetFormatter(f Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = f
}

// ─── StderrHandler ───────────────────────────────────────────────────────────

type StderrHandler struct {
	BaseHandler
	mu        sync.Mutex
	formatter Formatter
}

func NewStderrHandler() *StderrHandler {
	return &StderrHandler{
		BaseHandler: NewBaseHandler(LevelWarn, FormatJSON),
		formatter:   NewJSONFormatter(),
	}
}

func (h *StderrHandler) Handle(_ context.Context, entry Entry) error {
	if !h.Enabled(entry.Level) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	output, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, output)
	return nil
}

func (h *StderrHandler) Close() error { return nil }

// ─── NullHandler ─────────────────────────────────────────────────────────────

type NullHandler struct{ BaseHandler }

func NewNullHandler() *NullHandler {
	return &NullHandler{BaseHandler: NewBaseHandler(LevelDebug, FormatJSON)}
}

func (h *NullHandler) Handle(_ context.Context, _ Entry) error { return nil }
func (h *NullHandler) Close() error                            { return nil }

// ─── MultiHandler ─────────────────────────────────────────────────────────────

type MultiHandler struct{ handlers []Handler }

func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Handle(ctx context.Context, entry Entry) error {
	var errs []string
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, entry); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("handler errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (h *MultiHandler) Close() error {
	var errs []string
	for _, handler := range h.handlers {
		if err := handler.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (h *MultiHandler) Enabled(level Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) SetLevel(level Level) {
	for _, handler := range h.handlers {
		handler.SetLevel(level)
	}
}
