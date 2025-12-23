package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Setup initializes the application logger
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

