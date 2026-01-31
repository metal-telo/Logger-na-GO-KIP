# Многоэтапная сборка для оптимизации размера образа

# Этап 1: Сборка приложения
FROM golang:1.21-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копирование go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

# Этап 2: Финальный образ
FROM alpine:latest

# Установка CA сертификатов и wget для health check
RUN apk --no-cache add ca-certificates tzdata wget

# Создание пользователя для запуска приложения (безопасность)
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копирование бинарника из builder
COPY --from=builder /app/server .

# Создание директорий для логов и метрик
RUN mkdir -p /app/logs /app/metrics && \
    chown -R appuser:appuser /app

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Переменные окружения по умолчанию
ENV PORT=:8080
ENV LOG_LEVEL=info
ENV HTTP_LOGGING=true
ENV TRACING_ENABLED=false
ENV SERVICE_NAME=employee-management-system

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Запуск приложения
CMD ["./server"]

