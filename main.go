package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"employee-management-system/internal/handler"
	"employee-management-system/internal/repository"
	"employee-management-system/internal/service"
	"employee-management-system/internal/telemetry"
)

func main() {
	// Настройка JSON логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Запуск системы управления сотрудниками")

	// Инициализация метрик
	telemetry.InitMetrics()

	// Инициализация сервисов
	repo := repository.NewMemoryRepository()
	employeeService := service.NewEmployeeService(repo)
	h := handler.NewHandler(employeeService)

	// Запуск сервера
	server := h.StartServer(":8080")

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Завершение работы сервера...")
	server.Shutdown()
	slog.Info("Сервер остановлен")
}