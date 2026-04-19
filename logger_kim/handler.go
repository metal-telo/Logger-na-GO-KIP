package logger_kim

import "context"

// Handler is the interface for log output handlers.
// Implement this interface to create any custom output destination.
type Handler interface {
	Handle(ctx context.Context, entry Entry) error
	Close() error
	Enabled(level Level) bool
	SetLevel(level Level)
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	minLevel Level
	format   Format
}

func NewBaseHandler(level Level, format Format) BaseHandler {
	return BaseHandler{minLevel: level, format: format}
}

func (h *BaseHandler) Enabled(level Level) bool {
	return level >= h.minLevel
}

func (h *BaseHandler) SetLevel(level Level) {
	h.minLevel = level
}

func (h *BaseHandler) GetFormat() Format {
	return h.format
}
