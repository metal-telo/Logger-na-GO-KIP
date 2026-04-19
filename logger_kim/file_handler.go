package logger_kim

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ─── FileHandler с ротацией по размеру ──────────────────────────────────────

type FileHandler struct {
	BaseHandler
	mu          sync.Mutex
	file        *os.File
	filePath    string
	formatter   Formatter
	maxSize     int64
	maxBackups  int
	currentSize int64
}

type FileHandlerConfig struct {
	FilePath   string
	MaxSize    int64 // в MB
	MaxBackups int
	Format     Format
	Level      Level
}

func NewFileHandler(filePath string) (*FileHandler, error) {
	return NewFileHandlerWithConfig(FileHandlerConfig{
		FilePath:   filePath,
		MaxSize:    100,
		MaxBackups: 5,
		Format:     FormatJSON,
		Level:      LevelDebug,
	})
}

func NewFileHandlerWithConfig(cfg FileHandlerConfig) (*FileHandler, error) {
	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	var f Formatter
	switch cfg.Format {
	case FormatText:
		f = NewTextFormatter()
	case FormatPretty:
		f = NewPrettyFormatter()
	default:
		f = NewJSONFormatter()
	}

	return &FileHandler{
		BaseHandler: NewBaseHandler(cfg.Level, cfg.Format),
		file:        file,
		filePath:    cfg.FilePath,
		formatter:   f,
		maxSize:     cfg.MaxSize * 1024 * 1024,
		maxBackups:  cfg.MaxBackups,
		currentSize: stat.Size(),
	}, nil
}

func (h *FileHandler) Handle(_ context.Context, entry Entry) error {
	if !h.Enabled(entry.Level) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.maxSize > 0 && h.currentSize >= h.maxSize {
		if err := h.rotate(); err != nil {
			return fmt.Errorf("failed to rotate: %w", err)
		}
	}

	output, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	output += "\n"
	n, err := h.file.WriteString(output)
	if err != nil {
		return err
	}
	h.currentSize += int64(n)
	return nil
}

func (h *FileHandler) rotate() error {
	h.file.Close()
	for i := h.maxBackups - 1; i > 0; i-- {
		old := fmt.Sprintf("%s.%d", h.filePath, i)
		new := fmt.Sprintf("%s.%d", h.filePath, i+1)
		if _, err := os.Stat(old); err == nil {
			os.Remove(new)
			os.Rename(old, new)
		}
	}
	os.Rename(h.filePath, h.filePath+".1")
	file, err := os.OpenFile(h.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	h.file = file
	h.currentSize = 0
	return nil
}

func (h *FileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

// ─── DailyRotatingFileHandler ────────────────────────────────────────────────

type DailyRotatingFileHandler struct {
	BaseHandler
	mu           sync.Mutex
	file         *os.File
	baseFilePath string
	formatter    Formatter
	currentDate  string
	maxDays      int
}

func NewDailyRotatingFileHandler(baseFilePath string, maxDays int) (*DailyRotatingFileHandler, error) {
	h := &DailyRotatingFileHandler{
		BaseHandler:  NewBaseHandler(LevelDebug, FormatJSON),
		baseFilePath: baseFilePath,
		formatter:    NewJSONFormatter(),
		maxDays:      maxDays,
	}
	if err := h.openTodaysFile(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *DailyRotatingFileHandler) openTodaysFile() error {
	today := time.Now().Format("2006-01-02")
	if h.currentDate == today && h.file != nil {
		return nil
	}
	if h.file != nil {
		h.file.Close()
	}
	dir := filepath.Dir(h.baseFilePath)
	base := filepath.Base(h.baseFilePath)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]
	filePath := filepath.Join(dir, fmt.Sprintf("%s-%s%s", nameWithoutExt, today, ext))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	h.file = file
	h.currentDate = today
	return nil
}

func (h *DailyRotatingFileHandler) Handle(_ context.Context, entry Entry) error {
	if !h.Enabled(entry.Level) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.openTodaysFile()
	output, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.file.WriteString(output + "\n")
	return err
}

func (h *DailyRotatingFileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

func (h *DailyRotatingFileHandler) SetLevel(level Level) { h.BaseHandler.SetLevel(level) }
