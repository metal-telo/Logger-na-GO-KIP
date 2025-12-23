package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"employee-management/internal/handler"
	"employee-management/internal/logger"
	"employee-management/internal/repository"
	"employee-management/internal/service"
	"employee-management/internal/telemetry"
)

//go:embed static/*
var staticFiles embed.FS

const (
	LogFile     = "app.log"
	MetricsFile = "metrics.log"
	LogDir      = "logs"
	MetricsDir  = "metrics"
	ServerPort  = ":8080"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup logger
	logFile, err := logger.Setup(LogDir, LogFile)
	if err != nil {
		return fmt.Errorf("ошибка настройки логгера: %w", err)
	}
	defer logFile.Close()

	slog.Info("Логгер инициализирован", "log_file", LogDir+"/"+LogFile)

	// Setup metrics writer
	metricsFile, err := telemetry.SetupMetricsWriter(MetricsDir, MetricsFile)
	if err != nil {
		slog.Error("Ошибка настройки записи метрик", "error", err)
	} else {
		defer metricsFile.Close()
		slog.Info("Запись метрик в файл инициализирована", "metrics_file", MetricsDir+"/"+MetricsFile)
		go telemetry.StartMetricsWriter()
	}

	slog.Info("Трассировка отключена - Jaeger не запущен")
	telemetry.InitMetrics()

	// Initialize dependencies
	repo := repository.NewMemoryRepository()
	svc := service.NewEmployeeService(repo)
	h := handler.NewHandler(svc, staticFiles)

	// Create server
	server := &http.Server{
		Addr:         ServerPort,
		Handler:      h.InitRoutes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("Запуск сервера", "port", ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка запуска сервера", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Завершение работы сервера...")
	telemetry.WriteMetricsToFile()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка завершения работы сервера: %w", err)
	}

	slog.Info("Сервер остановлен")
	return nil
}

